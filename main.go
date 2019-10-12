package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly"
)

const (
	initStep    int = 0
	loginStep   int = 1
	tokenPage   int = 2
	createToken int = 3

	defaultSettingsPath = ".config/Code/User/settings.json"
)

var username string
var password string
var settings string

func init() {
	flag.StringVar(&username, "username", "", "evernote username")
	flag.StringVar(&password, "password", "", "evernote password")
	flag.StringVar(&settings, "settings", "", fmt.Sprintf("evernote settings file path, default ~/%s", defaultSettingsPath))
}

func resetEvermonkeyToken(settingsPath string, newToken string) {
	bytes, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		panic(err)
	}
	oldSet := string(bytes)

	old := `"evermonkey.token": ".*",`
	new := fmt.Sprintf(`"evermonkey.token": "%s",`, newToken)

	re3, _ := regexp.Compile(old)
	newSet := re3.ReplaceAllString(oldSet, new)

	// save settings
	outputFile, outputError := os.OpenFile(settingsPath, os.O_WRONLY|os.O_CREATE, 0666)
	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputWriter.WriteString(newSet)
	outputWriter.Flush()
}

func main() {
	flag.Parse()

	if username == "" || password == "" {
		fmt.Println("Usage:./ever-token-vscode -h")
		return
	}
	if settings == "" {
		settings = os.ExpandEnv(filepath.Join("$HOME", defaultSettingsPath))
		fmt.Println("use defaultSettingsPath", settings)
	}

	c := colly.NewCollector()

	// authenticate
	loginData := make(map[string]string)
	loginData["username"] = username
	loginData["password"] = password
	loginData["rememberMe"] = "true"
	loginData["login"] = "登录"
	loginData["hpts"] = ""
	loginData["hptsh"] = "="
	loginData["analyticsLoginOrigin"] = "login_action"
	loginData["clipperFlow"] = "false"
	loginData["showSwitchService"] = "true"
	loginData["usernameImmutable"] = "false"
	loginData["targetUrl"] = "/api/DeveloperToken.action"
	loginData["clipperFlow"] = "false"
	loginData["_sourcePage"] = ""
	loginData["__fp"] = ""

	step := initStep

	err := c.Post("https://app.yinxiang.com/Login.action", loginData)
	if err != nil {
		log.Fatal(err)
	}

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		// fmt.Println(r.StatusCode)
		if step == initStep {
			doc, _ := htmlquery.Parse(strings.NewReader(string(r.Body)))
			xmlNodes := htmlquery.Find(doc, "/html/body/div/div/div/script")

			for _, xmlNode := range xmlNodes {
				xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)

				scriptCode := xmlElem.ChildText("/")
				lines := strings.Split(scriptCode, "\n")
				for _, line := range lines {
					keys := strings.Split(line, "\"")
					if len(keys) == 5 {
						if keys[1] == "hpts" {
							loginData["hpts"] = keys[3]
						}
						if keys[1] == "hptsh" {
							loginData["hptsh"] = keys[3]
						}
					}
				}
			}
			xmlNodes = htmlquery.Find(doc, "/html/body/div/div/div/div/div/div/div/div/form/div/input")
			for _, xmlNode := range xmlNodes {
				xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
				name := xmlElem.Attr("name")
				// fmt.Println(name)
				if name == "_sourcePage" {
					loginData["_sourcePage"] = xmlElem.Attr("value")
				} else if name == "__fp" {
					loginData["__fp"] = xmlElem.Attr("value")
				}
			}

			// fmt.Println(loginData)
			step = loginStep
			err := c.Post("https://app.yinxiang.com/Login.action", loginData)
			if err != nil {
				log.Fatal(err)
			}

		} else if step == loginStep {
			// fmt.Println(string(r.Body))
			step = tokenPage
			err := c.Post("https://app.yinxiang.com/api/DeveloperToken.action", nil)
			if err != nil {
				log.Fatal(err)
			}
		} else if step == tokenPage {
			// fmt.Println(string(r.Body))

			createData := make(map[string]string)
			createData["secret"] = ""
			createData["create"] = "Create+a+developer+token"
			createData["csrfBusterToken"] = ""
			createData["_sourcePage"] = ""
			createData["__fp"] = ""

			doc, _ := htmlquery.Parse(strings.NewReader(string(r.Body)))
			xmlNodes := htmlquery.Find(doc, "/html/body/div/div/div/form/input")

			for _, xmlNode := range xmlNodes {
				xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
				name := xmlElem.Attr("name")
				if name == "secret" {
					createData["secret"] = xmlElem.Attr("value")
				} else if name == "csrfBusterToken" {
					createData["csrfBusterToken"] = xmlElem.Attr("value")
				}
			}

			xmlNodes = htmlquery.Find(doc, "/html/body/div/div/div/form/div/input")

			for _, xmlNode := range xmlNodes {
				xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
				name := xmlElem.Attr("name")
				if name == "_sourcePage" {
					createData["_sourcePage"] = xmlElem.Attr("value")
				} else if name == "__fp" {
					createData["__fp"] = xmlElem.Attr("value")
				}
			}

			step = createToken
			err := c.Post("https://app.yinxiang.com/api/DeveloperToken.action", createData)
			if err != nil {
				log.Fatal(err)
			}

		} else if step == createToken {
			// fmt.Println(string(r.Body))
			doc, _ := htmlquery.Parse(strings.NewReader(string(r.Body)))
			xmlNode := htmlquery.FindOne(doc, "/html/body/div/div/div/form/div/div/div/input")
			xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
			accessToken := xmlElem.Attr("value")

			if accessToken == "https://app.yinxiang.com/shard/s12/notestore" {
				// need Revoke your developer token
				createData := make(map[string]string)
				createData["secret"] = ""
				createData["noteStoreUrl"] = "https://app.yinxiang.com/shard/s12/notestore"
				createData["remove"] = "Revoke your developer token"
				createData["csrfBusterToken"] = ""
				createData["_sourcePage"] = ""
				createData["__fp"] = ""

				doc, _ := htmlquery.Parse(strings.NewReader(string(r.Body)))
				xmlNodes := htmlquery.Find(doc, "/html/body/div/div/div/form/input")

				for _, xmlNode := range xmlNodes {
					xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
					name := xmlElem.Attr("name")
					if name == "secret" {
						createData["secret"] = xmlElem.Attr("value")
					} else if name == "csrfBusterToken" {
						createData["csrfBusterToken"] = xmlElem.Attr("value")
					}
				}

				xmlNodes = htmlquery.Find(doc, "/html/body/div/div/div/form/div/input")

				for _, xmlNode := range xmlNodes {
					xmlElem := colly.NewXMLElementFromHTMLNode(r, xmlNode)
					name := xmlElem.Attr("name")
					if name == "_sourcePage" {
						createData["_sourcePage"] = xmlElem.Attr("value")
					} else if name == "__fp" {
						createData["__fp"] = xmlElem.Attr("value")
					}
				}

				step = tokenPage
				err := c.Post("https://app.yinxiang.com/api/DeveloperToken.action", createData)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				fmt.Println(accessToken)
				resetEvermonkeyToken(settings, accessToken)
			}

		}
	})

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// e.Request.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://app.yinxiang.com/api/DeveloperToken.action")
}
