package main

import "github.com/lifei6671/goapollo/goapollo"

func main() {

	config := func(client *goapollo.ApolloClient) {
		client.Port = 8080
	}

	apolloConfig := func(client *goapollo.ApolloClient) {
		config := goapollo.NewApolloConfig("http://dev.config.xin.com/", "6e77bd897fe903ac")
		config.NamespaceName = "TEST1.nginx"

		client.AddApolloConfig(config)
	}

	client := goapollo.NewApolloClient(config, apolloConfig)
	client.Run()
}
