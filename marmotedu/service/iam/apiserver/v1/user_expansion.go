// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

// The UserExpansion interface allows manually adding extra methods to the UserInterface.
type UserExpansion interface { // PatchStatus modifies the status of an existing node. It returns the copy
	// of the node that the server returns, or an error.
	// PatchStatus(ctx context.Context, nodeName string, data []byte) (*v1.Node, error)
}

/*
// PatchStatus modifies the status of an existing node. It returns the copy of
// the node that the server returns, or an error.
func (c *nodes) PatchStatus(ctx context.Context, nodeName string, data []byte) (*v1.Node, error) {
	result := &v1.Node{}
	err := c.client.Patch(types.StrategicMergePatchType).
		Resource("nodes").
		Name(nodeName).
		SubResource("status").
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}
*/
