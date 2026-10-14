// SPDX-License-Identifier: Apache-2.0

package cmd

import "testing"

func TestDistinctApproverRejectsOwner(t *testing.T) {
	if err := distinctApprover("owner", "owner"); err == nil {
		t.Fatal("owner approved their own exception")
	}
	if err := distinctApprover("owner", ""); err == nil {
		t.Fatal("empty approver accepted")
	}
	if err := distinctApprover("owner", "approver"); err != nil {
		t.Fatal(err)
	}
}
