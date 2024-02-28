package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// don't assume working directory; create file in same directory as executable
	fullPath, err := filepath.Abs(getArg(0, true))
	if err != nil {
		log.Println(err)
	}

	file, err := os.OpenFile(fullPath+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Print(err)
		os.Exit(12)
	}
	log.SetOutput(file)

	_ = godotenv.Load(fullPath + ".env")

	user := getEnv("LEGO_MIAB_USER", true)
	pass := getEnv("LEGO_MIAB_PASS", true)
	host := getEnv("LEGO_MIAB_HOST", true)

	// Expecting: ./command <verb> <domain> <value>
	// (unsupported) LOGO "RAW" mode is: ./command <verb> -- <domain> <token> <key_auth>
	verb := getVerb(getArg(1, true))
	domain := getArg(2, true)
	if domain == "--" {
		log.Println("Called in unsupported 'RAW' mode")
		os.Exit(11)
	}
	// if <domain> ends in period (".") remove it
	domain = strings.TrimSuffix(domain, ".")
	value := getArg(3, false)

	log.Printf("Requested: %s, %s, %s", verb, domain, value)
	response := doRequest(verb, user, pass, host, domain, value)
	if !strings.HasPrefix(response, "updated") {
		//log.Fatalf("Update failed: %s", response)
		log.Printf("Update failed: %s", response)
		os.Exit(2)
	}
	log.Printf("Success: %s, %s, %s", verb, domain, value)

	os.Exit(0)
}

func getArg(index int, required bool) string {
	count := len(os.Args)
	if index > (count - 1) {
		if required {
			//log.Fatalf("Argument number %d required, but %d provided", index, count - 1)
			log.Printf("Argument number %d required, but %d provided", index, count-1)
			os.Exit(3)
		}
		return ""
	}

	value := os.Args[index]
	if required && (value == "") {
		//log.Fatalf("Argument number %d can not be empty", index)
		log.Printf("Argument number %d can not be empty", index)
		os.Exit(4)
	}
	return value
}

func getEnv(name string, required bool) string {
	value, exist := os.LookupEnv(name)
	if required && !exist {
		//log.Fatalf("Environment variable %s required; alternately, use '%s.env' file", name, getProgramName())
		log.Printf("Environment variable %s required; alternately, use '%s.env' file", name, getProgramName())
		os.Exit(5)
	}
	return value
}

func getVerb(arg string) string {
	switch arg {
	case "present":
		return "PUT"
	case "cleanup":
		return "DELETE"
	default:
		//log.Fatalf("Invalid LEGO verb %s ('present' or 'cleanup' required)", arg)
		log.Printf("Invalid LEGO verb %s ('present' or 'cleanup' required)", arg)
		os.Exit(6)
	}
	return ""
}

func doRequest(verb string, user string, pass string, host string, domain string, value string) string {
	data := strings.NewReader(value)
	path := fmt.Sprintf("https://%s/admin/dns/custom/%s/TXT", host, domain)
	req, err := http.NewRequest(verb, path, data)
	if err != nil {
		//log.Fatalf("Error creating request: %s", err.Error())
		log.Printf("Error creating request: %s", err.Error())
		os.Exit(7)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//log.Fatalf("Error contacting DNS server: %s", err.Error())
		log.Printf("Error contacting DNS server: %s", err.Error())
		os.Exit(8)
	}

	//defer resp.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			//log.Fatalf("Error closing request: %s", err.Error())
			log.Printf("Error closing request: %s", err.Error())
			os.Exit(9)
		}
	}(resp.Body)

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		//log.Fatalf("Error reading response: %s", err.Error())
		log.Printf("Error reading response: %s", err.Error())
		os.Exit(10)
	}
	//fmt.Printf("%s", bodyText)
	return string(bodyText[:])
}

func getProgramName() string {
	return filepath.Base(getArg(0, true))
}
