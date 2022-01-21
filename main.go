package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/tuanna7593/alpies-bot/client"
	"github.com/tuanna7593/alpies-bot/config"
	"github.com/tuanna7593/alpies-bot/contracts"
	"github.com/tuanna7593/alpies-bot/templates"
)

var (
	TradeSignature     = []byte("Trade(address,uint256,address,address,uint256,uint256,bool)")
	TradeSignatureHash = crypto.Keccak256Hash(TradeSignature)
	BotToken           = "2143465691:AAEuKfYh3ZOZptSVXOJ5Ob5VbgrbmMs61nQ"
	ChannelId          = "@alpies_lover"
)

type Trade struct {
	Collection common.Address `json:"collection"`
	TokenID    *big.Int       `json:"token_id"`
	Seller     common.Address `json:"seller"`
	Buyer      common.Address `json:"buyer"`
	Tx         common.Hash    `json:"tx"`
	AskPrice   *big.Int       `json:"ask_price"`
	NetPrice   *big.Int       `json:"net_price"`
	WithBNB    bool           `json:"with_bnb"`
}

func main() {
	bot, err := tgbot.NewBotAPI(BotToken)
	if err != nil {
		log.Panic(err)
	}

	// read config
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	config.InitConfig(&cfg)

	// init ether client
	err = initEthClient()
	if err != nil {
		log.Fatal(err)
	}

	c := client.GetEtheClient()
	marketContractAbi, err := abi.JSON(strings.NewReader(string(contracts.PCSNFTMarketABI.ABI)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("cfg.NFTMarketAddress: ", cfg.NFTMarketAddress)
	filterLogs := ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress(cfg.NFTMarketAddress),
		},
		Topics: [][]common.Hash{
			{common.HexToHash("0xdaac0e40b8f01e970d08d4e8ae57ac31a5845fffde104c43f05e19bbec78491e")},
		},
	}

	logsChan := make(chan types.Log)
	subscription, err := c.SubscribeFilterLogs(context.Background(), filterLogs, logsChan)
	if err != nil {
		log.Fatal("failed to subscribe pcs nft market log:", err)
	}

	for {
		select {
		case data := <-logsChan:
			fmt.Printf("data.Topics lengh:%d\n", len(data.Topics))
			var tradeLog Trade
			err := marketContractAbi.UnpackIntoInterface(&tradeLog, "Trade", data.Data)
			if err != nil {
				fmt.Printf("failed to unpack data to trade log struct:%v\n", err)
				continue
			}
			tradeLog.Collection = common.HexToAddress(data.Topics[1].Hex())
			tradeLog.TokenID = data.Topics[2].Big()
			tradeLog.Seller = common.HexToAddress(string(data.Topics[3].Hex()))

			fmt.Printf("trade log:%+v\n", tradeLog)
			p := make(tgbot.Params)
			p.AddNonEmpty("chat_id", ChannelId)
			p.AddNonEmpty("parse_mode", tgbot.ModeMarkdown)
			p.AddNonEmpty("text", templates.PCS_ALPIE_BUY_TMPL)
			resp, err := bot.MakeRequest("sendMessage", p)
			if err != nil {
				fmt.Printf("failed to post message to telegram:%v\n", err)
				continue
			}
			if !resp.Ok {
				fmt.Printf("resp failed:%+v\n", resp)
			}
		case errSubscription := <-subscription.Err():
			fmt.Println("websocket closed")
			log.Fatal(errSubscription)
		}
	}
}

func loadConfig() (cfg config.ClientConfig, err error) {
	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}

func initEthClient() error {
	cfg := config.GetConfig()
	if cfg.RPCEndpoint == "" {
		return fmt.Errorf("not found rpc endpoint")
	}

	cli, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return err
	}

	client.SetEtheClient(cli)
	return nil
}
