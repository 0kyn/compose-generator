package cmd

import (
	"compose-generator/util"
	"os"
	"strings"

	dcu "github.com/compose-generator/dcu"
	dcu_model "github.com/compose-generator/dcu/model"
)

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// Remove services from an existing compose file
func Remove(serviceNames []string, flagRun bool, flagDetached bool, flagWithVolumes bool, flagForce bool, flagAdvanced bool) {
	util.ClearScreen()

	// Ask for custom YAML file
	path := "./docker-compose.yml"
	if flagAdvanced {
		path = util.TextQuestionWithDefault("From which compose file do you want to remove a service?", "./docker-compose.yml")
	}

	removeFromFile(path, serviceNames, flagWithVolumes, flagForce, flagAdvanced)

	// Run if the corresponding flag is set
	if flagRun || flagDetached {
		util.DockerComposeUp(flagDetached)
	}
}

// --------------------------------------------------------------- Private functions ---------------------------------------------------------------

func removeFromFile(filePath string, serviceNames []string, flagWithVolumes bool, flagForce bool, flagAdvanced bool) {
	util.P("Parsing compose file ... ")
	// Load compose file
	composeFile, err := dcu.DeserializeFromFile(filePath)
	if err != nil {
		util.Error("Internal error - unable to parse compose file", err, true)
	}
	util.Done()

	// Ask for service(s)
	if len(serviceNames) == 0 {
		var items []string
		for k := range composeFile.Services {
			items = append(items, k)
		}
		serviceNames = util.MultiSelectMenuQuestion("Which services do you want to remove?", items)
		util.Pel()
	}

	for _, serviceName := range serviceNames {
		// Remove volumes
		if flagWithVolumes {
			removeVolumesForService(composeFile, serviceName, flagForce)
		}

		// Remove service
		util.P("Removing service '" + serviceName + "' ... ")
		var networkCount = make(map[string]int32)
		delete(composeFile.Services, serviceName) // Remove service itself
		for k, s := range composeFile.Services {
			s.DependsOn = util.RemoveStringFromSlice(s.DependsOn, serviceName) // Remove dependencies on service
			s.Links = util.RemoveStringFromSlice(s.Links, serviceName)         // Remove links on service
			for networkName := range composeFile.Networks {                    // Collect count of every network
				if util.SliceContainsString(s.Networks, networkName) {
					networkCount[networkName]++
				}
			}
			composeFile.Services[k] = s
		}

		// Remove unused networks
		for networkName := range composeFile.Networks {
			if networkCount[networkName] < 2 {
				delete(composeFile.Networks, networkName) // Delete network itself
				for k, s := range composeFile.Services {  // Delete references on service
					s.Networks = util.RemoveStringFromSlice(s.Networks, networkName)
					composeFile.Services[k] = s
				}
			}
		}
		util.Done()
	}

	// Write to file
	util.P("Saving compose file ... ")
	if err := dcu.SerializeToFile(composeFile, "./docker-compose.yml"); err != nil {
		util.Error("Could not write yaml to compose file", err, true)
	}
	util.Done()
}

func removeVolumesForService(composeFile dcu_model.ComposeFile, serviceName string, flagForce bool) {
	reallyDeleteVolumes := true
	if !flagForce {
		reallyDeleteVolumes = util.YesNoQuestion("Do you really want to delete all attached volumes of service '"+serviceName+"'. All data will be lost.", false)
	}
	if reallyDeleteVolumes {
		util.P("Removing volumes of '" + serviceName + "' ... ")
		volumes := composeFile.Services[serviceName].Volumes
		for _, paths := range volumes {
			path := paths
			if strings.Contains(path, ":") {
				path = path[:strings.IndexByte(path, ':')]
			}
			// Check if volume is used by another container
			canBeDeleted := true
		out:
			for k, s := range composeFile.Services {
				if k != serviceName {
					for _, pathsInner := range s.Volumes {
						pathInner := pathsInner
						if strings.Contains(pathInner, ":") {
							pathInner = pathInner[:strings.IndexByte(pathInner, ':')]
						}
						if pathInner == path {
							canBeDeleted = false
							break out
						}
					}
				}
			}
			if canBeDeleted && util.FileExists(path) {
				os.RemoveAll(path)
			}
		}
		util.Done()
	}
}