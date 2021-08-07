// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ory/ladon"

	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/marmotedu/component-base/pkg/util/homedir"

	"github.com/marmotedu/marmotedu-sdk-go/marmotedu"
	"github.com/marmotedu/marmotedu-sdk-go/tools/clientcmd"
)

func main() {
	var iamconfig *string
	if home := homedir.HomeDir(); home != "" {
		iamconfig = flag.String(
			"iamconfig",
			filepath.Join(home, ".iam", "config"),
			"(optional) absolute path to the iamconfig file",
		)
	} else {
		iamconfig = flag.String("iamconfig", "", "absolute path to the iamconfig file")
	}
	flag.Parse()

	// use the current context in iamconfig
	config, err := clientcmd.BuildConfigFromFlags("", *iamconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := marmotedu.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	request := &ladon.Request{
		Resource: "resources:articles:ladon-introduction",
		Action:   "delete",
		Subject:  "users:peter",
		Context: ladon.Context{
			"remoteIP": "192.168.0.5",
		},
	}

	// Authorize the request
	fmt.Println("Authorize request...")
	ret, err := clientset.Iam().AuthzV1().Authz().Authorize(context.TODO(), request, metav1.AuthorizeOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Authorize response: %s.\n", ret.ToString())
}
