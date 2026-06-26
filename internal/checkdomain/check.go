package checkdomain

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"zabbix-domain-expiry/internal/output"
	"zabbix-domain-expiry/internal/rdap"
	"zabbix-domain-expiry/internal/whois"
)

type Options struct {
	Domain            string
	Warning           int
	Critical          int
	RDAPServer        string
	WhoisServer       string
	Debug             bool
}

func Run(opts Options) {
	debug := func(msg string) {
		if opts.Debug {
			fmt.Fprintf(os.Stderr, "INFO [%s]: %s\n", time.Now().Format("15:04:05"), msg)
		}
	}

	if opts.Domain == "" {
		output.Exit(output.StateUnknown, 0, "", "State: UNKNOWN ; No domain specified")
	}

	rdapClient := rdap.NewClient(debug)
	rdapClient.Override = opts.RDAPServer

	expiration, rdapErr := rdapClient.GetExpiration(opts.Domain)
	if rdapErr != nil {
		debug(fmt.Sprintf("RDAP lookup failed: %v, falling back to WHOIS", rdapErr))
	}

	if expiration == "" {
		debug(fmt.Sprintf("Falling back to WHOIS for %s", opts.Domain))

		whoisClient := whois.NewClient(debug)
		whoisClient.Override = opts.WhoisServer

		response, err := whoisClient.Query(opts.Domain)
		if err != nil {
			output.Exit(output.StateUnknown, 0, "", fmt.Sprintf("State: UNKNOWN ; %s", err))
		}

		expiration, err = whois.ParseExpiration(response)
		if err != nil {
			output.Exit(output.StateUnknown, 0, "", fmt.Sprintf("State: UNKNOWN ; %s for %s", err, opts.Domain))
		}
	}

	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(expiration) {
		output.Exit(output.StateUnknown, 0, "", fmt.Sprintf("State: UNKNOWN ; Invalid expiration date format: %s", expiration))
	}

	expTime, err := time.Parse("2006-01-02", expiration)
	if err != nil {
		output.Exit(output.StateUnknown, 0, "", fmt.Sprintf("State: UNKNOWN ; Failed to parse expiration date: %s", expiration))
	}

	now := time.Now().UTC()
	expDays := int(expTime.UTC().Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)

	debug(fmt.Sprintf("Days left: %d ; Exp date: %s", expDays, expiration))

	if expDays >= 0 {
		if expDays <= opts.Critical {
			output.Exit(output.StateCritical, expDays, expiration,
				fmt.Sprintf("State: CRITICAL ; Days left: %d ; Expire date: %s", expDays, expiration))
		}
		if expDays <= opts.Warning {
			output.Exit(output.StateWarning, expDays, expiration,
				fmt.Sprintf("State: WARNING ; Days left: %d ; Expire date: %s", expDays, expiration))
		}
		output.Exit(output.StateOK, expDays, expiration,
			fmt.Sprintf("State: OK ; Days left: %d ; Expire date: %s", expDays, expiration))
	}

	daysSince := -expDays
	output.Exit(output.StateCritical, expDays, expiration,
		fmt.Sprintf("State: CRITICAL ; Days since expired: %d ; Expire date: %s", daysSince, expiration))
}

func ParseArgs(args []string) (Options, error) {
	opts := Options{
		Warning:  30,
		Critical: 7,
	}

	args = preprocessArgs(args)
	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		case "-V", "--version":
			output.PrintVersion()
			os.Exit(0)
		case "-d", "--domain":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing domain value")
			}
			i++
			opts.Domain = args[i]
		case "-w", "--warning":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					i++
					opts.Warning = v
					break
				}
			}
		case "-c", "--critical":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					i++
					opts.Critical = v
					break
				}
			}
		case "-P", "--path":
			if i+1 < len(args) {
				i++
			}
		case "-s", "--whois-server":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing whois server value")
			}
			i++
			opts.WhoisServer = args[i]
			if !isValidServer(opts.WhoisServer) {
				return opts, fmt.Errorf("invalid WHOIS server: %s", opts.WhoisServer)
			}
		case "-r", "--rdap-server":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing rdap server value")
			}
			i++
			opts.RDAPServer = args[i]
			if !isValidURL(opts.RDAPServer) {
				return opts, fmt.Errorf("invalid RDAP server: %s", opts.RDAPServer)
			}
		case "-z", "--debug":
			opts.Debug = true
		default:
			return opts, fmt.Errorf("invalid argument: %s", arg)
		}
		i++
	}

	return opts, nil
}

func preprocessArgs(args []string) []string {
	var result []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-r", "--rdap-server", "-s", "--whois-server":
			result = append(result, args[i])
			if i+1 < len(args) {
				if args[i+1] == "" {
					result = append(result, "0")
					i++
				} else {
					i++
					result = append(result, args[i])
				}
			}
		default:
			result = append(result, args[i])
		}
	}
	return result
}

var (
	urlPattern    = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(/.*)?$|^0$`)
	serverPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$|^0$`)
)

func isValidURL(value string) bool {
	return urlPattern.MatchString(value)
}

func isValidServer(value string) bool {
	return serverPattern.MatchString(value)
}

func printHelp() {
	fmt.Printf(`check_domain - v%s

This program checks the expiration status of a domain using RDAP or WHOIS protocols.
No external binaries required — all networking and parsing is built in.

Usage: check_domain -h | -d <domain> [-c <critical>] [-w <warning>] [-s <whois_server>] [-r <rdap_server>] [-z]
Options:
-h|--help            Print detailed help
-V|--version         Print version information
-d|--domain          Domain name to check
-w|--warning         Warning threshold (days, default: 30)
-c|--critical        Critical threshold (days, default: 7)
-s|--whois-server    Specific WHOIS server (use "" for default lookup)
-r|--rdap-server     Specific RDAP server URL (use "" for IANA lookup)
-z|--debug           Enable debug output
`, output.Version)
}