// SPDX-License-Identifier: Apache-2.0

package collection

import "k8s.io/apimachinery/pkg/api/resource"

func resourceQuantity(value string) resource.Quantity { return resource.MustParse(value) }
