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

	usersClient := iamclient.APIV1().Users()

	user := &v1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Nickname: "sdkexample",
		Password: "Sdk@2020",
		Email:    "user@qq.com",
		Phone:    "1812884xxxx",
	}

	// Create user
	fmt.Println("Creating user...")
	ret, err := usersClient.Create(context.TODO(), user, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created user %q.\n", ret.GetObjectMeta().GetName())
	defer func() {
		// Delete secret
		fmt.Println("Deleting user...")
		if err := usersClient.Delete(context.TODO(), "sdk", metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Delete user failed: %s\n", err.Error())
			return
		}
		fmt.Println("Deleted user.")
	}()

	// Get user
	prompt()
	fmt.Println("Geting user...")
	ret, err = usersClient.Get(context.TODO(), user.Name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Get user %q.\n", ret.GetObjectMeta().GetName())

	// Update user
	prompt()
	fmt.Println("Updating user...")
	user = &v1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Nickname: "sdkexample_update",
		Email:    "user_update@qq.com",
		Phone:    "1812885xxxx",
	}
	ret, err = usersClient.Update(context.TODO(), user, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Updated user...")
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
