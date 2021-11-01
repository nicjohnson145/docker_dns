package main

import (
	"fmt"
	"os"
	"time"
)

type Provider interface {
	Load(Settings)
	Reconcile([]string) error
}

func ExitErr(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func GetProvider(name string) Provider {
	switch name {
	case cloudflare:
		x := &Cloudflare{}
		return x
	default:
		ExitErr(fmt.Sprintf("Unknown provider %v", name))
	}
	// Appease the compiler, it doesn't know ExitErr terminates the program
	panic("Unreachable code")
}

func main() {
	settings := LoadSettings()
	provider := GetProvider(settings.ProviderName)
	provider.Load(settings)

	for {
		expectedDomains, err := getExpectedSubdomains(settings.Domain)
		if err != nil {
			fmt.Printf("error getting expected domains: %v", err)
		}
		err = provider.Reconcile(expectedDomains)
		if err != nil {
			fmt.Printf("error reconciling: %v", err)
		}

		time.Sleep(5 * time.Second)
	}
}
