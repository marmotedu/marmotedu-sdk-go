// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/marmotedu/api/apiserver/v1"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/marmotedu/component-base/pkg/util/homedir"

	"github.com/marmotedu/marmotedu-sdk-go/marmotedu/service/iam"
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

	// create the iamclient
	iamclient, err := iam.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secretsClient := iamclient.APIV1().Secrets()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Expires:     3724075800,
		Description: "test secret for sdk",
	}

	// Create secret
	fmt.Println("Creating secret...")
	ret, err := secretsClient.Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created secret %q.\n", ret.GetObjectMeta().GetName())
	defer func() {
		// Delete secret
		fmt.Println("Deleting secret...")
		if err := secretsClient.Delete(context.TODO(), "sdk", metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Delete secret failed: %s\n", err.Error())
			return
		}
		fmt.Println("Deleted secret.")
	}()

	// Get secret
	prompt()
	fmt.Println("Geting secret...")
	ret, err = secretsClient.Get(context.TODO(), "sdk", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Get secret %q.\n", ret.GetObjectMeta().GetName())

	// Update secret
	prompt()
	fmt.Println("Updating secret...")
	secret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Expires:     4071231000,
		Description: "test secret for sdk_update",
	}
	ret, err = secretsClient.Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Updated secret..., new expires: %d\n", ret.Expires)

	// List secrets
	prompt()
	fmt.Println("Listing secrets...")
	list, err := secretsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (secretID: %s, secretKey: %s, expires: %d)\n",
			d.Name, d.SecretID, d.SecretKey, d.Expires)
	}
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
