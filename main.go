package main

import (
	"context"
	"fmt"
	"os"

	"github.com/opentofu/terraform-provider-testfunctions/internal/testfunctions"
)

func main() {
	ctx := context.Background()

	provider := testfunctions.NewProvider()
	err := provider.Serve(ctx)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error starting provider: %s\n", err))
		os.Exit(1)
	}
}
