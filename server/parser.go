package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func (config Config) toFastCfg() (fastCfg FastCfg) {
	fastCfg = FastCfg{}

	for _, service := range config.Services {

		// Only Prod
		if service.DevPort == 0 {

			if fastCfg[service.Port] == nil {
				fastCfg[service.Port] = FastPort{}
			}
			if fastCfg[service.Port][service.Host] == nil {
				fastCfg[service.Port][service.Host] = FastHost{}
			}

			// Passende App finden
			for _, app := range config.Apps {

				if app.Name == service.App {
					sort.Slice(app.Routes, func(i, j int) bool {
						if app.Routes[i].Path == "/" {
							return false
						}
						if app.Routes[j].Path == "/" {
							return true
						}

						return strings.Count(app.Routes[i].Path, "/") < strings.Count(app.Routes[j].Path, "/")
					})

					// Routen in FastCfg einfügen
					for _, route := range app.Routes {

						if len(route.Path) > 1 {
							route.Path = strings.TrimSuffix(route.Path, "/")
						}
						route.Endpoint = strings.TrimSuffix(route.Endpoint, "/")

						if service.Hosts != nil {
							for _, host := range service.Hosts {
								fastCfg[service.Port][host] = append(fastCfg[service.Port][host], FastRoute{route.Path, route.Endpoint})
							}
						} else {
							fastCfg[service.Port][service.Host] = append(fastCfg[service.Port][service.Host], FastRoute{route.Path, route.Endpoint})
						}
					}
				}
			}
		}
	}

	return
}

func (config Config) toDevCfgAndDevServerCfg() (devCfg DevCfg, devServerCfg DevServerCfg) {
	devCfg = DevCfg{}
	devServerCfg = DevServerCfg{}

	for _, service := range config.Services {

		// Only Dev
		if service.DevPort != 0 {

			for _, host := range append(service.Hosts, service.Host) {

				if host != "" {

					devService := DevService{
						Name:        service.Name,
						Description: service.Description,
						Host:        host,
						Port:        service.Port,
						DevPort:     service.DevPort,
						Routes:      []DevRoute{},
					}

					// Passende App finden
					for _, app := range config.Apps {

						if app.Name == service.App {
							sort.Slice(app.Routes, func(i, j int) bool {
								if app.Routes[i].Path == "/" {
									return false
								}
								if app.Routes[j].Path == "/" {
									return true
								}

								return strings.Count(app.Routes[i].Path, "/") < strings.Count(app.Routes[j].Path, "/")
							})

							// Routen in FastCfg einfügen
							for _, route := range app.Routes {

								if len(route.Path) > 1 {
									route.Path = strings.TrimSuffix(route.Path, "/")
								}
								route.Endpoint = strings.TrimSuffix(route.Endpoint, "/")

								devService.Routes = append(devService.Routes, DevRoute{route.Name, route.Description, route.Path, route.Endpoint, append([]string{}, route.Endpoint)})
							}
						}
					}

					devCfg[devService.Port] = append(devCfg[devService.Port], &devService)
					devServerCfg[devService.DevPort] = append(devServerCfg[devService.DevPort], &devService)
				}
			}
		}
	}

	return
}

func loadConfig() (config Config) {
	if len(os.Args[1:]) == 0 {
		fmt.Println("Keine Konfiguration angegeben")
		os.Exit(1)
	}

	filePath, _ := filepath.Abs(os.Args[1:][0])
	yamlFile, err := ioutil.ReadFile(filePath)

	if err != nil {
		fmt.Println("Konfiguration konnte nicht geöffnet werden")
		os.Exit(1)
	}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		fmt.Println("Konfiguration ist fehlerhaft")
		os.Exit(1)
	}

	return
}
