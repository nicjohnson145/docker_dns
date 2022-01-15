package main

import (
	"fmt"
	"os"
	"time"
	"flag"
	"log"
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
	var oneTime = flag.Bool("one-time", false, "Execute 1 loop and exit")
	flag.Parse()

	settings := LoadSettings()
	provider := GetProvider(settings.ProviderName)
	provider.Load(settings)

	for {
		_ = reconcile(settings, provider)
		if *oneTime {
			break
		}
		time.Sleep(time.Duration(settings.Interval) * time.Second)
	}
}

func reconcile(settings Settings, provider Provider) error {
	expectedDomains, err := getExpectedSubdomains(settings.Domain)
	if err != nil {
		log.Printf("error getting expected domains: %v\n", err)
		return err
	}
	err = provider.Reconcile(expectedDomains)
	if err != nil {
		log.Printf("error reconciling: %v\n", err)
		return err
	}

	return nil
}
