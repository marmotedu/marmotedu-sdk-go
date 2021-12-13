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

	"github.com/ory/ladon"

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

	policiesClient := iamclient.APIV1().Policies()

	var policyConditions = ladon.Conditions{
		"owner": &ladon.EqualsSubjectCondition{},
	}

	policy := &v1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Policy: v1.AuthzPolicy{
			DefaultPolicy: ladon.DefaultPolicy{
				Description: "description",
				Subjects:    []string{"user"},
				Effect:      ladon.AllowAccess,
				Resources:   []string{"articles:<[0-9]+>"},
				Actions:     []string{"create", "update"},
				Conditions:  policyConditions,
			},
		},
	}

	// Create policy
	fmt.Println("Creating policy...")
	ret, err := policiesClient.Create(context.TODO(), policy, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Created policy %q.\n", ret.GetObjectMeta().GetName())
	defer func() {
		// Delete policy
		fmt.Println("Deleting policy...")
		if err := policiesClient.Delete(context.TODO(), "sdk", metav1.DeleteOptions{Unscoped: true}); err != nil {
			fmt.Printf("Delete policy failed: %s\n", err.Error())
			return
		}
		fmt.Println("Deleted policy.")
	}()

	// Get policy
	prompt()
	fmt.Println("Geting policy...")
	ret, err = policiesClient.Get(context.TODO(), "sdk", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Get policy %q.\n", ret.GetObjectMeta().GetName())

	// Update policy
	prompt()
	fmt.Println("Updating policy...")
	policy = &v1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sdk",
		},
		Policy: v1.AuthzPolicy{
			DefaultPolicy: ladon.DefaultPolicy{
				Description: "description - update",
				Subjects:    []string{"user"},
				Effect:      ladon.AllowAccess,
				Resources:   []string{"articles:<[0-9]+>"},
				Actions:     []string{"create", "update"},
				Conditions:  policyConditions,
			},
		},
	}
	ret, err = policiesClient.Update(context.TODO(), policy, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Updated policy..., new policy: `%s`\n", ret.Policy.Description)

	// List policys
	prompt()
	fmt.Println("Listing policies...")
	list, err := policiesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (policy: `%s`)\n", d.Name, d.Policy.Description)
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
