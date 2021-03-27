package commands

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/fatih/color"
	yaml "gopkg.in/yaml.v3"

	"compose-generator/model"
	"compose-generator/parser"
	"compose-generator/utils"
)

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// Generate a docker compose configuration
func Generate(flagAdvanced bool, flagRun bool, flagDetached bool, flagForce bool, flagWithInstructions bool, flagWithDockerfile bool) {
	utils.ClearScreen()

	// Execute SafetyFileChecks
	if !flagForce {
		utils.ExecuteSafetyFileChecks(flagWithInstructions, flagWithDockerfile)
	}

	// Welcome Message
	utils.Heading("Welcome to Compose Generator!")
	utils.Pl("Please continue by answering a few questions:")
	utils.Pel()

	// Project name
	projectName := utils.TextQuestion("What is the name of your project:")
	if projectName == "" {
		utils.Error("Error. You must specify a project name!", true)
	}

	// Docker Swarm compatibility (default: no)
	//dockerSwarm := utils.YesNoQuestion("Should your compose file be used for distributed deployment with Docker Swarm?", false)
	//utils.Pl(dockerSwarm)

	// Predefined stack (default: yes)
	usePredefinedStack := utils.YesNoQuestion("Do you want to use a predefined stack?", true)
	if usePredefinedStack {
		generateFromPredefinedTemplate(projectName, flagAdvanced, flagWithInstructions, flagWithDockerfile)
	} else {
		generateFromScratch(projectName, flagAdvanced, flagForce)
	}

	// Run if the corresponding flag is set
	if flagRun || flagDetached {
		utils.DockerComposeUp(flagDetached)
	} else {
		// Print success message
		utils.Pel()
		utils.SuccessMessage("🎉 Done! You now can execute \"$ docker-compose up\" to launch your app! 🎉")
	}
}

// --------------------------------------------------------------- Private functions ---------------------------------------------------------------

func generateFromPredefinedTemplate(projectName string, flagAdvanced bool, flagWithInstructions bool, flagWithDockerfile bool) {
	utils.ClearScreen()

	// Load stacks from templates
	templateData := parser.ParsePredefinedServices()
	// Predefined stack menu
	var proxy_items []string
	var tls_helper_items []string
	var frontend_items []string
	var backend_items []string
	var database_items []string
	var db_admin_items []string
	for _, t := range templateData {
		switch t.Type {
		case "proxy":
			proxy_items = append(proxy_items, t.Label)
		case "tls-helper":
			tls_helper_items = append(tls_helper_items, t.Label)
		case "frontend":
			frontend_items = append(frontend_items, t.Label)
		case "backend":
			backend_items = append(backend_items, t.Label)
		case "database":
			database_items = append(database_items, t.Label)
		case "db-admin-tool":
			db_admin_items = append(db_admin_items, t.Label)
		}
	}

	// Ask for production environment
	//production := utils.YesNoQuestion("Are you generating this for a production environment?", false)

	// Ask for compose file version
	compose_version := "3.9"
	if flagAdvanced {
		/*compose_version = */ utils.TextQuestionWithDefault("Docker compose file version:", compose_version)
	}

	// Ask for proxies
	/*proxy_selection := */
	utils.MenuQuestion("Which proxy do you want to use?", proxy_items)

	// Ask for proxy tls helpers
	/*tls_helper_selections := */
	utils.MenuQuestion("Which TLS helper do you want to use?", tls_helper_items)

	// Ask for frontends
	/*frontend_selections := */
	utils.MultiSelectMenuQuestion("Which frontend framework do you want to use?", frontend_items)

	// Ask for backends
	/*backend_selections := */
	utils.MultiSelectMenuQuestion("Which backend framework do you want to use?", backend_items)

	// Ask for databases
	/*database_selections := */
	utils.MultiSelectMenuQuestion("Which database engine do you want to use?", database_items)

	// Ask for db admin tools
	/*db_admin_selections := */
	utils.MultiSelectMenuQuestion("Which database admin tool do you want to use?", db_admin_items)

	utils.Pel()

	// Ask configured questions to the user
	envMap := make(map[string]string)
	envMap["PROJECT_NAME"] = projectName
	envMap["PROJECT_NAME_CONTAINER"] = strings.ReplaceAll(strings.ToLower(projectName), " ", "-")

	/*for _, q := range templateData[index].Questions {
		if !q.Advanced || (q.Advanced && flagAdvanced) {
			switch q.Type {
			case 1: // Yes/No
				defaultValue, _ := strconv.ParseBool(q.DefaultValue)
				envMap[q.EnvVar] = strconv.FormatBool(utils.YesNoQuestion(q.Text, defaultValue))
			case 2: // Text
				if q.Validator != "" {
					var customValidator survey.Validator
					switch q.Validator {
					case "port":
						customValidator = utils.PortValidator
					default:
						customValidator = func(val interface{}) error {
							validate := validator.New()
							if validate.Var(val.(string), "required,"+q.Validator) != nil {
								return errors.New("please provide a valid input")
							}
							return nil
						}
					}
					envMap[q.EnvVar] = utils.TextQuestionWithDefaultAndValidator(q.Text, q.DefaultValue, customValidator)
				} else {
					envMap[q.EnvVar] = utils.TextQuestionWithDefault(q.Text, q.DefaultValue)
				}
			}
		} else {
			envMap[q.EnvVar] = q.DefaultValue
		}
	}

	// Ask for custom volume paths
	volumesMap := make(map[string]string)
	if len(templateData[index].Volumes) > 0 {
		for _, v := range templateData[index].Volumes {
			if !v.Advanced || (v.Advanced && flagAdvanced) {
				if !v.WithDockerfile || (v.WithDockerfile && flagWithDockerfile) {
					envMap[v.EnvVar] = utils.TextQuestionWithDefault(v.Text, v.DefaultValue)
				} else {
					envMap[v.EnvVar] = v.DefaultValue
				}
			} else {
				envMap[v.EnvVar] = v.DefaultValue
			}
			volumesMap[v.DefaultValue] = envMap[v.EnvVar]
		}
		utils.Pel()
	}

	// Copy template files
	fmt.Print("Copying predefined template ... ")
	srcPath := utils.GetPredefinedServicesPath() + "/" + templateData[index].Dir
	dstPath := "."

	var err error
	for _, f := range templateData[index].Files {
		switch f.Type {
		case "compose", "env":
			os.Remove(dstPath + "/" + f.Path)
			if utils.FileExists(srcPath + "/" + f.Path) {
				err = copy.Copy(srcPath+"/"+f.Path, dstPath+"/"+f.Path)
			}
		case "docs", "docker":
			if (flagWithInstructions && f.Type == "docs") || (flagWithDockerfile && f.Type == "docker") {
				os.Remove(dstPath + "/" + f.Path)
				if utils.FileExists(srcPath + "/" + f.Path) {
					err = copy.Copy(srcPath+"/"+f.Path, dstPath+"/"+f.Path)
				}
			}
		}
	}
	if err != nil {
		utils.Error("Could not copy predefined template files.", true)
	}
	utils.Done()

	// Create volumes
	fmt.Print("Creating volumes ... ")
	for src, dst := range volumesMap {
		os.RemoveAll(dst)
		src = srcPath + src[1:]

		opt := copy.Options{
			Skip: func(src string) (bool, error) {
				return strings.HasSuffix(src, ".gitkeep"), nil
			},
			OnDirExists: func(src string, dst string) copy.DirExistsAction {
				return copy.Replace
			},
		}
		err := copy.Copy(src, dst, opt)
		if err != nil {
			utils.Error("Could not copy volume files.", true)
		}
	}
	utils.Done()

	// Replace variables
	fmt.Print("Applying customizations ... ")
	utils.ReplaceVarsInFile("./docker-compose.yml", envMap)
	if utils.FileExists("./environment.env") {
		utils.ReplaceVarsInFile("./environment.env", envMap)
	}
	if flagWithDockerfile && utils.FileExists("./Dockerfile") {
		utils.ReplaceVarsInFile("./Dockerfile", envMap)
	}
	utils.Done()

	if utils.FileExists("./environment.env") {
		// Generate secrets
		fmt.Print("Generating secrets ... ")
		secretsMap := utils.GenerateSecrets("./environment.env", templateData[index].Secrets)
		utils.Done()
		// Print secrets to console
		utils.Pel()
		utils.Pl("Following secrets were automatically generated:")
		for key, secret := range secretsMap {
			fmt.Print("   " + key + ": ")
			color.Yellow(secret)
		}
	}*/
}

func generateFromScratch(projectName string, flagAdvanced bool, flagForce bool) {
	utils.ClearScreen()

	// Create custom stack
	utils.Heading("Okay. Let's create a custom stack for you!")
	utils.Pel()

	services := make(map[string]model.Service)
	i := 1
	for another := true; another; another = utils.YesNoQuestion("Generate another service?", true) {
		utils.ClearScreen()
		color.Blue("Service no. " + strconv.Itoa(i) + ":")
		service, serviceName, _ := AddService(services, flagAdvanced, flagForce, true)
		services[serviceName] = service
		utils.ClearScreen()
		color.Green("✓ Created Service '" + serviceName + "'")
		i++
	}
	utils.ClearScreen()

	// Ask for the dependencies
	var serviceNames []string
	for name := range services {
		serviceNames = append(serviceNames, name)
	}
	for serviceName, service := range services {
		currentService := service
		service.DependsOn = utils.MultiSelectMenuQuestion("On which services should your service '"+serviceName+"' depend?", utils.RemoveStringFromSlice(serviceNames, serviceName))
		services[serviceName] = currentService
	}

	// Generate compose file
	utils.P("Generating compose file ... ")
	composeFile := model.ComposeFile{}
	composeFile.Version = "3.9"
	composeFile.Services = services
	utils.Done()

	// Write to file
	utils.P("Saving compose file ... ")
	output, err1 := yaml.Marshal(&composeFile)
	err2 := ioutil.WriteFile("./docker-compose.yml", output, 0777)
	if err1 != nil || err2 != nil {
		utils.Error("Could not write yaml to compose file.", true)
	}
	utils.Done()
}
