package whois

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var monthAbbr = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var monthFull = map[string]int{
	"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6,
	"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
}

type dateHandler func(line string) (string, bool)

func ParseExpiration(data string) (string, error) {
	lines := strings.Split(data, "\n")
	handlers := expirationHandlers()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, handler := range handlers {
			if date, ok := handler(line); ok {
				if err := validateDate(date); err != nil {
					continue
				}
				return date, nil
			}
		}
	}

	return "", fmt.Errorf("no expiration date found")
}

func validateDate(date string) error {
	date = strings.TrimSpace(date)
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date) {
		return fmt.Errorf("invalid date format: %s", date)
	}
	return nil
}

func formatYMD(year, month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func mon2moy(month string) int {
	month = strings.ToLower(month)
	if v, ok := monthAbbr[month]; ok {
		return v
	}
	if len(month) >= 3 {
		if v, ok := monthAbbr[month[:3]]; ok {
			return v
		}
	}
	return 0
}

func month2moy(month string) int {
	month = strings.ToLower(month)
	if v, ok := monthFull[month]; ok {
		return v
	}
	return mon2moy(month)
}

func expirationHandlers() []dateHandler {
	return []dateHandler{
		handleExpiresMonthDayYear,
		handleExpiryDateDDMonYYYY,
		handleExpireDateISO,
		handleExpiryDateDDMMYYYYSlash,
		handleExpiryDateYYYYMMDDSlash,
		handleExpirationDateISO,
		handleExpirationDateDot,
		handleExpiresYYYYMMDD,
		handleExpiresOnBracket,
		handleRecordExpiresOn,
		handleExpiryDateISO,
		handleDomainExpirationDateWeekday,
		handleExpiresDDMonYYYY,
		handleExpirationDateDotTime,
		handleExpiresISO,
		handleExpireDateTime,
		handleExpireDateTimeShort,
		handleRenewalDateDot,
		handlePaidTillDot,
		handlePaidTillISO,
		handleValidUntil,
		handleExpireDot,
		handleDomainDateBilledUntil,
		handleRegistrarRegistrationExpiration,
		handleExpirationDateDDMMYYYYLabel,
		handleDomainExpirationDateGMT,
		handleExpirationDateUTC,
		handleExpirationDateTimeNoTZ,
		handleExpiryDateUTC,
		handleExpiresOnUTCOffset,
		handleDateSlash,
		handleExpiryDateSlash,
		handleValidDateOrExpires,
		handleStateBracket,
		handleExpiresAt,
		handleRenewalDateISO,
		handleExpiryDateDDMMYYYYDash,
		handleValidity,
		handleExpiredOn,
		handleExpiresGeneric,
		handleExpiresDotTime,
		handleExpiresTime,
		handleExpiresTimeShort,
		handleExpiryDateLower,
		handleExpirationDateSpace,
		handleExpirationTime,
		handleBilledUntil,
		handleRenewalFollowup,
	}
}

var (
	reExpiresMonthDayYear      = regexp.MustCompile(`(?i)expires:\s+([A-Za-z]+)\s+(\d+)\s+(\d{4})`)
	reExpiryDateDDMonYYYY      = regexp.MustCompile(`(?i)[Ee]xpiry[Dd]ate:.*(\d{2})-([A-Za-z]{3})-(\d{4})`)
	reExpireDateISO            = regexp.MustCompile(`(?i)[Ee]xpire[- ]date:.*(\d{4}-\d{2}-\d{2})`)
	reExpiryDateDDMMYYYYSlash  = regexp.MustCompile(`(?i)[Ee]xpiry date:.*(\d{2})/(\d{2})/(\d{4})`)
	reExpiryDateYYYYMMDDSlash  = regexp.MustCompile(`(?i)[Ee]xpiry date:.*(\d{4})/(\d{2})/(\d{2})`)
	reExpirationDateISO        = regexp.MustCompile(`(?i)Expiration Date:.*(\d{4}-\d{2}-\d{2})T`)
	reExpirationDateDot        = regexp.MustCompile(`(?i)Expiration [Dd]ate:.*(\d{2})\.(\d{2})\.(\d{4})`)
	reExpiresYYYYMMDD          = regexp.MustCompile(`(?i)expires:.*(\d{8})`)
	reExpiresOnBracket         = regexp.MustCompile(`(?i)\[Expires on\].*(\d{4})/(\d{2})/(\d{2})`)
	reRecordExpiresOn          = regexp.MustCompile(`(?i)Record expires on\s+(\d{4}-\d{2}-\d{2})`)
	reExpiryDateISO            = regexp.MustCompile(`(?i)Expiry Date:\s+(\d{4}-\d{2}-\d{2})`)
	reDomainExpirationWeekday  = regexp.MustCompile(`(?i)Domain Expiration Date:\s+[A-Za-z]{3}\s+([A-Za-z]{3})\s+(\d{2})\s+\d{2}:\d{2}:\d{2}\s+(\d{4})`)
	reExpiresDDMonYYYY         = regexp.MustCompile(`(?i)expires:.*(\d{2})-([A-Za-z]{3})-(\d{4})`)
	reExpirationDateDotTime    = regexp.MustCompile(`(?i)Expiration date:.*(\d{2})\.(\d{2})\.(\d{4})\s+\d{2}:\d{2}`)
	reExpiresISO               = regexp.MustCompile(`(?i)expires:.*(\d{4}-\d{2}-\d{2})`)
	reExpireDateTime           = regexp.MustCompile(`(?i)expire:.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}:\d{2}`)
	reExpireDateTimeShort      = regexp.MustCompile(`(?i)expire:.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}`)
	reRenewalDateDot           = regexp.MustCompile(`(?i)renewal date:.*(\d{4})\.(\d{2})\.(\d{2})`)
	rePaidTillDot              = regexp.MustCompile(`(?i)paid-till:.*(\d{4})\.(\d{2})\.(\d{2})`)
	rePaidTillISO              = regexp.MustCompile(`(?i)paid-till:\s*(\d{4}-\d{2}-\d{2})`)
	reValidUntil               = regexp.MustCompile(`(?i)Valid Until:.*(\d{4}-\d{2}-\d{2})`)
	reExpireDot                = regexp.MustCompile(`(?i)expire:.*(\d{2})\.(\d{2})\.(\d{4})`)
	reDomainDateBilledUntil    = regexp.MustCompile(`(?i)domain_datebilleduntil:.*(\d{4}-\d{2}-\d{2})T`)
	reRegistrarExpiration      = regexp.MustCompile(`(?i)Registrar Registration Expiration Date:.*(\d{4}-\d{2}-\d{2})T`)
	reExpirationDDMMYYYYLabel  = regexp.MustCompile(`(?i)Expiration Date.*\(dd/mm/yyyy\):\s*(\d{2})/(\d{2})/(\d{4})`)
	reDomainExpirationGMT      = regexp.MustCompile(`(?i)Domain Expiration Date:.*[A-Za-z]{3}\s+([A-Za-z]{3})\s+(\d{2})\s+\d{2}:\d{2}:\d{2}\s+GMT\s+(\d{4})`)
	reExpirationDateUTC        = regexp.MustCompile(`(?i)Expiration Date:\s*(\d{2})-([A-Za-z]{3})-(\d{4})\s+\d{2}:\d{2}:\d{2}\s+UTC`)
	reExpirationDateTimeNoTZ   = regexp.MustCompile(`(?i)Expiration Date:\s*(\d{2})-([A-Za-z]{3})-(\d{4})\s+\d{2}:\d{2}:\d{2}`)
	reExpiryDateUTC            = regexp.MustCompile(`(?i)Expiry Date:\s*(\d{2})\s+([A-Za-z]{3})\s+(\d{4})\s+\d{2}:\d{2}:\d{2}\s+UTC`)
	reExpiresOnUTCOffset       = regexp.MustCompile(`(?i)expires on.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}:\d{2}\s+\(UTC\+`)
	reDateSlash                = regexp.MustCompile(`(?i)date:.*(\d{4})/(\d{2})/(\d{2})`)
	reExpiryDateSlash          = regexp.MustCompile(`(?i)Expiry Date:.*(\d{2})/(\d{2})/(\d{4})`)
	reValidDateOrExpires       = regexp.MustCompile(`(?i)(Valid-date|Expir(es|y)):.*(\d{4}-\d{2}-\d{2})`)
	reStateBracket             = regexp.MustCompile(`(?i)\[State\].*(\d{4})/(\d{2})/(\d{2})`)
	reExpiresAt                = regexp.MustCompile(`(?i)expires at:.*(\d{2})/(\d{2})/(\d{4})`)
	reRenewalDateISO           = regexp.MustCompile(`(?i)Renewal Date:.*(\d{4}-\d{2}-\d{2})`)
	reExpiryDateDDMMYYYYDash   = regexp.MustCompile(`(?i)Expiry Date:.*(\d{2})-(\d{2})-(\d{4})`)
	reValidity                 = regexp.MustCompile(`(?i)validity:\s*(\S+)`)
	reExpiredOn                = regexp.MustCompile(`(?i)Expired on:.*(\d{4}-\d{2}-\d{2})`)
	reExpiresGeneric           = regexp.MustCompile(`(?i)Expires.*(\d{4}-\d{2}-\d{2})`)
	reExpiresDotTime           = regexp.MustCompile(`(?i)^expires\.?:\s*(\d{1,2})\.(\d{1,2})\.(\d{4})\s+\d{2}:\d{2}:\d{2}`)
	reExpiresTime              = regexp.MustCompile(`(?i)expires:.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}:\d{2}`)
	reExpiresTimeShort         = regexp.MustCompile(`(?i)expires:.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}`)
	reExpiryDateLower          = regexp.MustCompile(`(?i)Expiry date:.*(\d{2})-([A-Za-z]{3})-(\d{4})`)
	reExpirationDateSpace      = regexp.MustCompile(`(?i)Expiration Date:\s+(\d{4}-\d{2}-\d{2})\s`)
	reExpirationTime           = regexp.MustCompile(`(?i)Expiration Time:.*(\d{4}-\d{2}-\d{2})\s+\d{2}:\d{2}:\d{2}`)
	reBilledUntil              = regexp.MustCompile(`(?i)billed\s*until:\s*(\d{4}-\d{2}-\d{2})T`)
	reRenewalFollowup          = regexp.MustCompile(`(?i)renewal.*(\d{2})\s+([A-Za-z]{3})\s+(\d{4})`)
)

func handleExpiresMonthDayYear(line string) (string, bool) {
	m := reExpiresMonthDayYear.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := month2moy(m[1])
	day, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpiryDateDDMonYYYY(line string) (string, bool) {
	m := reExpiryDateDDMonYYYY.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpireDateISO(line string) (string, bool) {
	m := reExpireDateISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiryDateDDMMYYYYSlash(line string) (string, bool) {
	m := reExpiryDateDDMMYYYYSlash.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiryDateYYYYMMDDSlash(line string) (string, bool) {
	m := reExpiryDateYYYYMMDDSlash.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpirationDateISO(line string) (string, bool) {
	m := reExpirationDateISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpirationDateDot(line string) (string, bool) {
	m := reExpirationDateDot.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiresYYYYMMDD(line string) (string, bool) {
	m := reExpiresYYYYMMDD.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	raw := m[1]
	return fmt.Sprintf("%s-%s-%s", raw[0:4], raw[4:6], raw[6:8]), true
}

func handleExpiresOnBracket(line string) (string, bool) {
	m := reExpiresOnBracket.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleRecordExpiresOn(line string) (string, bool) {
	m := reRecordExpiresOn.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiryDateISO(line string) (string, bool) {
	m := reExpiryDateISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleDomainExpirationDateWeekday(line string) (string, bool) {
	m := reDomainExpirationWeekday.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[1])
	day, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpiresDDMonYYYY(line string) (string, bool) {
	m := reExpiresDDMonYYYY.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpirationDateDotTime(line string) (string, bool) {
	m := reExpirationDateDotTime.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiresISO(line string) (string, bool) {
	m := reExpiresISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpireDateTime(line string) (string, bool) {
	m := reExpireDateTime.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpireDateTimeShort(line string) (string, bool) {
	m := reExpireDateTimeShort.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleRenewalDateDot(line string) (string, bool) {
	m := reRenewalDateDot.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handlePaidTillDot(line string) (string, bool) {
	m := rePaidTillDot.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handlePaidTillISO(line string) (string, bool) {
	m := rePaidTillISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleValidUntil(line string) (string, bool) {
	m := reValidUntil.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpireDot(line string) (string, bool) {
	m := reExpireDot.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleDomainDateBilledUntil(line string) (string, bool) {
	m := reDomainDateBilledUntil.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleRegistrarRegistrationExpiration(line string) (string, bool) {
	m := reRegistrarExpiration.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpirationDateDDMMYYYYLabel(line string) (string, bool) {
	m := reExpirationDDMMYYYYLabel.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleDomainExpirationDateGMT(line string) (string, bool) {
	m := reDomainExpirationGMT.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[1])
	day, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpirationDateUTC(line string) (string, bool) {
	m := reExpirationDateUTC.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpirationDateTimeNoTZ(line string) (string, bool) {
	m := reExpirationDateTimeNoTZ.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpiryDateUTC(line string) (string, bool) {
	m := reExpiryDateUTC.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpiresOnUTCOffset(line string) (string, bool) {
	m := reExpiresOnUTCOffset.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleDateSlash(line string) (string, bool) {
	m := reDateSlash.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiryDateSlash(line string) (string, bool) {
	m := reExpiryDateSlash.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleValidDateOrExpires(line string) (string, bool) {
	m := reValidDateOrExpires.FindStringSubmatch(line)
	if len(m) != 3 {
		return "", false
	}
	return m[2], true
}

func handleStateBracket(line string) (string, bool) {
	m := reStateBracket.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiresAt(line string) (string, bool) {
	m := reExpiresAt.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleRenewalDateISO(line string) (string, bool) {
	m := reRenewalDateISO.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiryDateDDMMYYYYDash(line string) (string, bool) {
	m := reExpiryDateDDMMYYYYDash.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleValidity(line string) (string, bool) {
	m := reValidity.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	if strings.EqualFold(m[1], "N/A") {
		return "2100-01-01", true
	}
	parts := strings.Split(m[1], "-")
	if len(parts) != 3 {
		return "", false
	}
	day, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	year, _ := strconv.Atoi(parts[2])
	return formatYMD(year, month, day), true
}

func handleExpiredOn(line string) (string, bool) {
	m := reExpiredOn.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiresGeneric(line string) (string, bool) {
	m := reExpiresGeneric.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiresDotTime(line string) (string, bool) {
	m := reExpiresDotTime.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	return formatYMD(year, month, day), true
}

func handleExpiresTime(line string) (string, bool) {
	m := reExpiresTime.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiresTimeShort(line string) (string, bool) {
	m := reExpiresTimeShort.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpiryDateLower(line string) (string, bool) {
	m := reExpiryDateLower.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}

func handleExpirationDateSpace(line string) (string, bool) {
	m := reExpirationDateSpace.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleExpirationTime(line string) (string, bool) {
	m := reExpirationTime.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleBilledUntil(line string) (string, bool) {
	m := reBilledUntil.FindStringSubmatch(line)
	if len(m) != 2 {
		return "", false
	}
	return m[1], true
}

func handleRenewalFollowup(line string) (string, bool) {
	m := reRenewalFollowup.FindStringSubmatch(line)
	if len(m) != 4 {
		return "", false
	}
	month := mon2moy(m[2])
	day, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi(m[3])
	if month == 0 {
		return "", false
	}
	return formatYMD(year, month, day), true
}