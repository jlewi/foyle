---
description: Adding Level 1 Assertions
title: Adding Assertions
weight: 20
---

## Adding Level 1 Evals

1. Define the Assertion in [eval/assertions.go](../../../../../app/pkg/eval/assertions.go)
2. Update [Assertor in assertor.go](../app/pkg/eval/assertor.go) to include the new assertion
3. Update [AssertRow proto](../protos/eval/eval.proto) to include the new assertion
4. Update [toAssertionRow](../../../../../app/pkg/eval/service.go) to include the new assertion in `As