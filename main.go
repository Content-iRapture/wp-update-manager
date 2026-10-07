package main

import (
	"os/exec"
	"fmt"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)
const version = "0.1.1"
const domain = "elliott.irapture.com"

type Plugin struct {
	Name string `json:"name"`
	Update any `json:"update"`
	Version string `json:"version"`
	UpdateVersion string `json:"update_version"`
}


func main() {
	fmt.Printf("%s\n", version)
	cmd := exec.Command("plesk", "bin", "pleskbackup", "--domains-name", domain, "--incremental")	
	out, err := cmd.Output()
	
	if (err != nil) {
		fmt.Printf("Error: %v", err)
	} else {
		fmt.Printf("Out: %v", out)
	}
	
	cmd = exec.Command("plesk", "ext", "wp-toolkit", "--wp-cli", "-instance-id", "400", "--", "plugin", "list", "--format=json", "--fields=name,status,update,version,update_version")
	out, err = cmd.Output()

	if (err != nil) {
		log.Fatalf("Error: %v\n", err)
	}
	
	var plugins []Plugin
	reader := bytes.NewReader(out)
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&plugins)

	if (err != nil) {
		log.Fatalf("Error: %v\n", err)
	}
	
	fmt.Println("------- Plugin List --------")
	for _, plugin := range plugins {
		fmt.Printf("Name: '%s'\n\t?Update: '%s'\n\tVersion: '%s'\n\tUpdate Version: '%s'\n\tUpdate Available: '%t'\n", 
					plugin.Name,
					plugin.Update,
					plugin.Version,
					plugin.UpdateVersion,
					plugin.UpdateAvailable(),
		)
		
		// UpdateAvailable & UpdateUnavailable are not exhaustive
		// !UpdateAvailable() & !UpdateUnavailable can be true
		// therefore !UpdateAvailable() does not imply UpdateUnavailable()
		if (plugin.UpdateAvailable()) {
			cmd := exec.Command("plesk", "ext", "wp-toolkit", "--wp-cli", "-instance-id", "400", "--", "plugin", "update", plugin.Name)
			out, err := cmd.Output()

			if (err != nil) {
				// TODO: this should rollback
				err2 := restore(domain)
				if err2 != nil {
					log.Fatalf("Error durring restore: %v\n\t- occured during error: %v",
						err2,
						err,
					)
				}
				log.Fatalf("Error: %v\n", err)
			}

			res, err := http.Get(fmt.Sprintf("https://%s/", domain))

			if (err != nil) {
				// TODO: this should rollback
				err2 := restore(domain)
				if err2 != nil {
					log.Fatalf("Error durring restore: %v\n\t- occured during error: %v",
						err2,
						err,
					)
				}
				log.Fatalf("Error: %v\n", err)
			}

			if (res.StatusCode != http.StatusOK) { 
				err2 := restore(domain)
				if err2 != nil {
					log.Fatalf("Error durring restore: %v\n\t- occured during error: %v",
						err2,
						err,
					)
				}
				log.Fatalf("Error: %s\nRestored.\n", res.Status)
			}

			fmt.Printf("\tPLUGIN UPDATE TEXT: %s\n", out)
		}
	}
	fmt.Println("----- Plugin List End ------")

	
}

func restore(d string) error {

	restore := exec.Command(
		"sh", 
		"-c",
		`plesk bin pleskrestore --restore "$(ls -t /var/lib/psa/dumps/domains/"$1"/*.xml | head -1)" -level domains`,
		"sh", 
		domain,
	)

	if err := restore.Run(); err != nil {
		return err
	}

	return nil
}

func (p Plugin) UpdateAvailable() bool {
	s, _ := p.Update.(string)
	return s == "available"
}

/* For later
func (p Plugin) UpdateUnavailable() bool {
	s, _ := p.Update.(string)
	return s == "unavailable"
}
*/
