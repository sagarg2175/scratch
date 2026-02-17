package handler

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	trans    ut.Translator
)

type errorValidResponse struct {
	Success bool     `json:"success" example:"false"`
	Message []string `json:"message" example:"Error Message"`
	Errorno []string `json:"errornno"`
}

func newErrorValidResponse(message []string, errorno []string) errorValidResponse {
	return errorValidResponse{
		Success: false,
		Message: message,
		Errorno: errorno,
	}
}

type errordbResponse struct {
	Success bool     `json:"success" example:"false"`
	Message []string `json:"message" example:"Error message"`
	Errorno []string `json:"errorno"`
}

// newErrorResponse is a helper function to create an error response body
func newErrordbResponse(message []string, errorno []string) errordbResponse {
	return errordbResponse{
		Success: false,
		Message: message,
		Errorno: errorno,
	}
}

func myvalidate(f1 validator.FieldLevel) bool {
	fieldvalue := f1.Field().Int()
	return fieldvalue == 10
}

func customNameValidator(f1 validator.FieldLevel) bool {
	name, ok := f1.Field().Interface().(string)
	if !ok {
		return false
	}

	allowedNames := []string{"Admin", "Manager", "Supervisor"}
	for _, n := range allowedNames {
		if n == name {
			return true
		}
	}
	return false

}

type customTranslator struct {
	ut.Translator
}

func (ct *customTranslator) Add(translationId, format string, override bool) error {
	return ct.Translator.Add(translationId, format, override)
}

func (ct *customTranslator) T(translationID string, args ...interface{}) string {

	var stringArgs []string
	for _, arg := range args {
		if strArg, ok := arg.(string); ok {
			stringArgs = append(stringArgs, strArg)
		}
	}

	translated, _ := ct.Translator.T(translationID, stringArgs...)
	return translated
}

func trns1() {
	en := en.New()
	uni = ut.New(en, en)
	trans, _ = uni.GetTranslator("en")
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		return name
	})
	validate.RegisterValidation("myvalidate", myvalidate)
	validate.RegisterValidation("customName", customNameValidator)

	validate.RegisterTranslation("customName", trans, func(ut ut.Translator) error {
		return ut.Add("customName", "must be one of: Admin, Manager, Supervisor", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("customName", fe.Field())
		return t
	})

	en_translations.RegisterDefaultTranslations(validate, trans)
}

func handleError(ctx *gin.Context, err error) {
	var errormsg string
	statusCode, ok := errorStatusMap[err]
	if !ok {
		re := regexp.MustCompile(`cannot unmarshal (.*?) into Go struct field (.*?) of type (.*)$`)
		matches := re.FindStringSubmatch(err.Error())
		re1 := regexp.MustCompile(`invalid character '(.+?)'`)
		matches1 := re1.FindStringSubmatch(err.Error())
		if len(matches) == 4 {
			expectedType := matches[3]
			fieldarray := strings.Split(matches[2], ".")
			fieldvalue := fieldarray[1]
			errormsg = "Send " + expectedType + " for field: " + fieldvalue
		} else if len(matches1) == 2 {
			errormsg = "Malformed json request"
		}
		statusCode = http.StatusUnprocessableEntity
	}

	var errRsp errorValidResponse
	var errorMessages []string
	var erronumbers []string
	erronumbers = append(erronumbers, "UP1")
	if errormsg == "" {
		errorMessages = append(errorMessages, err.Error())
		errRsp = newErrorValidResponse(errorMessages, erronumbers)
	} else {
		//errormsgs:= newErrorResponse(errormsg)
		errorMessages = append(errorMessages, errormsg)
		errRsp = newErrorValidResponse(errorMessages, erronumbers)
	}

	ctx.JSON(statusCode, errRsp)
}

func handleValidation(ctx *gin.Context, s interface{}) bool {

	trns1() // initialize validator + translations + register customName.

	var errorMessages []string
	var erronumbers []string
	err := validate.Struct(s)

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {

			fieldName := e.StructField()
			// Access user-defined tag value using reflection
			t := reflect.TypeOf(s)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			structField, _ := t.FieldByName(fieldName)
			if e.Tag() == "myvalidate" {
				errorMessages = append(errorMessages, e.Field()+" must be equal to  10")
				erronumbers = append(erronumbers, "CST1")
			} else {
				//errorMessages = append(errorMessages, e.Translate(trans)+" Field: "+structField.Tag.Get("json"))
				errorMessages = append(errorMessages, e.Translate(trans))
				userDefinedValue := structField.Tag.Get("u")
				erronumbers = append(erronumbers, tagToNumber[e.Tag()]+userDefinedValue)
			}
		}
		errRsp := newErrorValidResponse(errorMessages, erronumbers)
		ctx.JSON(http.StatusBadRequest, errRsp)
		return false
	}
	return true
}

func handledbError(ctx *gin.Context, err error) {

	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
	}

	sqlStateRegex := regexp.MustCompile(`SQLSTATE (\d+)`)
	e := err.Error()
	matches := sqlStateRegex.FindStringSubmatch(e)
	var sqlState string
	if len(matches) >= 2 {
		sqlState = matches[1]
	} else {
		estring := err.Error()
		errRsps := newErrordbResponse([]string{estring}, []string{"POTH02"})
		ctx.JSON(statusCode, errRsps)
		return
	}

	var errdbslice []string
	var errdbno []string
	errordbClass1 := errordbMap[sqlState]
	
	dberror := strings.Split(errordbClass1, "—")
	errdbslice = append(errdbslice, dberror[1])
	errdbno = append(errdbno, dberror[0])
	errRsp := newErrordbResponse(errdbslice, errdbno)

	ctx.JSON(statusCode, errRsp)
}

var tagToNumber = map[string]string{
	"eqcsfield":                     "F1",
	"eqfield":                       "F2",
	"fieldcontains":                 "F3",
	"fieldexcludes":                 "F4",
	"gtcsfield":                     "F5",
	"gtecsfield":                    "F6",
	"gtefield":                      "F7",
	"gtfield":                       "F8",
	"ltcsfield":                     "F9",
	"ltecsfield":                    "F10",
	"ltefield":                      "F11",
	"ltfield":                       "F12",
	"necsfield":                     "F13",
	"nefield":                       "F14",
	"cidr":                          "N1",
	"cidrv4":                        "N2",
	"cidrv6":                        "N3",
	"datauri":                       "N4",
	"fqdn":                          "N5",
	"hostname":                      "N6",
	"hostname_port":                 "N7",
	"hostname_rfc1123":              "N8",
	"ip":                            "N9",
	"ip4_addr":                      "N10",
	"ip6_addr":                      "N11",
	"ip_addr":                       "N12",
	"ipv4":                          "N13",
	"ipv6":                          "N14",
	"mac":                           "N15",
	"tcp4_addr":                     "N16",
	"tcp6_addr":                     "N17",
	"tcp_addr":                      "N18",
	"udp4_addr":                     "N19",
	"udp6_addr":                     "N20",
	"udp_addr":                      "N21",
	"unix_addr":                     "N22",
	"uri":                           "N23",
	"url":                           "N24",
	"http_url":                      "N25",
	"url_encoded":                   "N26",
	"urn_rfc2141":                   "N27",
	"alpha":                         "S1",
	"alphanum":                      "S2",
	"alphanumunicode":               "S3",
	"alphaunicode":                  "S4",
	"ascii":                         "S5",
	"boolean":                       "S6",
	"contains":                      "S7",
	"containsany":                   "S8",
	"containsrune":                  "S9",
	"endsnotwith":                   "S10",
	"endswith":                      "S11",
	"excludes":                      "S12",
	"excludesall":                   "S13",
	"excludesrune":                  "S14",
	"lowercase":                     "S15",
	"multibyte":                     "S16",
	"number":                        "S17",
	"numeric":                       "S18",
	"printascii":                    "S19",
	"startsnotwith":                 "S20",
	"startswith":                    "S21",
	"uppercase":                     "S22",
	"base64":                        "FMT1",
	"base64url":                     "FMT2",
	"base64rawurl":                  "FMT3",
	"bic":                           "FMT4",
	"bcp47_language_tag":            "FMT5",
	"btc_addr":                      "FMT6",
	"btc_addr_bech32":               "FMT7",
	"credit_card":                   "FMT8",
	"mongodb":                       "FMT9",
	"cron":                          "FMT10",
	"spicedb":                       "FMT11",
	"datetime":                      "FMT12",
	"e164":                          "FMT13",
	"email":                         "FMT14",
	"eth_addr":                      "FMT15",
	"hexadecimal":                   "FMT16",
	"hexcolor":                      "FMT17",
	"hsl":                           "FMT18",
	"hsla":                          "FMT19",
	"html":                          "FMT20",
	"html_encoded":                  "FMT21",
	"isbn":                          "FMT22",
	"isbn10":                        "FMT23",
	"isbn13":                        "FMT24",
	"issn":                          "FMT25",
	"iso3166_1_alpha2":              "FMT26",
	"iso3166_1_alpha3":              "FMT27",
	"iso3166_1_alpha_numeric":       "FMT28",
	"iso3166_2":                     "FMT29",
	"iso4217":                       "FMT30",
	"json":                          "FMT31",
	"jwt":                           "FMT32",
	"latitude":                      "FMT33",
	"longitude":                     "FMT34",
	"luhn_checksum":                 "FMT35",
	"postcode_iso3166_alpha2":       "FMT36",
	"postcode_iso3166_alpha2_field": "FMT37",
	"rgb":                           "FMT38",
	"rgba":                          "FMT39",
	"ssn":                           "FMT40",
	"timezone":                      "FMT41",
	"uuid":                          "FMT42",
	"uuid3":                         "FMT43",
	"uuid3_rfc4122":                 "FMT44",
	"uuid4":                         "FMT45",
	"uuid4_rfc4122":                 "FMT46",
	"uuid5":                         "FMT47",
	"uuid5_rfc4122":                 "FMT48",
	"uuid_rfc4122":                  "FMT49",
	"md4":                           "FMT50",
	"md5":                           "FMT51",
	"sha256":                        "FMT52",
	"sha384":                        "FMT53",
	"sha512":                        "FMT54",
	"ripemd128":                     "FMT55",

	"tiger128":             "FMT57",
	"tiger160":             "FMT58",
	"tiger192":             "FMT59",
	"semver":               "FMT60",
	"ulid":                 "FMT61",
	"cve":                  "FMT62",
	"eq":                   "C1",
	"eq_ignore_case":       "C2",
	"gt":                   "C3",
	"gte":                  "C4",
	"lt":                   "C5",
	"lte":                  "C6",
	"ne":                   "C7",
	"ne_ignore_case":       "C8",
	"dir":                  "O1",
	"dirpath":              "O2",
	"file":                 "O3",
	"filepath":             "O4",
	"image":                "O5",
	"isdefault":            "O6",
	"len":                  "O7",
	"max":                  "O8",
	"min":                  "O9",
	"oneof":                "O10",
	"required":             "O11",
	"required_if":          "O12",
	"required_unless":      "O13",
	"required_with":        "O14",
	"required_with_all":    "O15",
	"required_without":     "O16",
	"required_without_all": "O17",
	"excluded_if":          "O18",
	"excluded_unless":      "O19",
	"excluded_with":        "O20",
	"excluded_with_all":    "O21",
	"excluded_without":     "O22",
	"excluded_without_all": "O23",
	"unique":               "O24",
	"iscolor":              "A1",
	"country_code":         "A2",
}

var errordbMap = map[string]string{

	"03000": "03—SQL Statement Not Yet Complete",
	"08000": "08—Connection Exception",
	"08003": "08—Connection Exception",
	"08006": "08—Connection Exception",
	"08001": "08—Connection Exception",
	"08004": "08—Connection Exception",
	"08007": "08—Connection Exception",
	"08P01": "08—Connection Exception",
	"09000": "09—Triggered Action Exception",
	"0A000": "0A—Feature Not Supported",
	"0B000": "0B—Invalid Transaction Initiation",
	"0F000": "0F—Locator Exception",
	"0F001": "0F—Locator Exception",
	"0L000": "0L—Invalid Grantor",
	"0LP01": "0L—Invalid Grantor",
	"0P000": "0P—Invalid Role Specification",
	"0Z000": "0Z—Diagnostics Exception",
	"0Z002": "0Z—Diagnostics Exception",
	"20000": "20—Case Not Found",
	"21000": "21—Cardinality Violation",
	"22000": "22—Data Exception",
	"2202E": "22—Data Exception",
	"22021": "22—Data Exception",
	"22008": "22—Data Exception",
	"22012": "22—Data Exception",
	"22005": "22—Data Exception",
	"2200B": "22—Data Exception",
	"22022": "22—Data Exception",
	"22015": "22—Data Exception",
	"2201E": "22—Data Exception",
	"22014": "22—Data Exception",
	"22016": "22—Data Exception",
	"2201F": "22—Data Exception",
	"2201G": "22—Data Exception",
	"22018": "22—Data Exception",
	"22007": "22—Data Exception",
	"22019": "22—Data Exception",
	"2200D": "22—Data Exception",
	"22025": "22—Data Exception",
	"22P06": "22—Data Exception",
	"22010": "22—Data Exception",
	"22023": "22—Data Exception",
	"22013": "22—Data Exception",
	"2201B": "22—Data Exception",
	"2201W": "22—Data Exception",
	"2201X": "22—Data Exception",
	"2202H": "22—Data Exception",
	"2202G": "22—Data Exception",
	"22009": "22—Data Exception",
	"2200C": "22—Data Exception",
	"2200G": "22—Data Exception",
	"22004": "22—Data Exception",
	"22002": "22—Data Exception",
	"22003": "22—Data Exception",
	"2200H": "22—Data Exception",
	"22026": "22—Data Exception",
	"22001": "22—Data Exception",
	"22011": "22—Data Exception",
	"22027": "22—Data Exception",
	"22024": "22—Data Exception",
	"2200F": "22—Data Exception",
	"22P01": "22—Data Exception",
	"22P02": "22—Data Exception",
	"22P03": "22—Data Exception",
	"22P04": "22—Data Exception",
	"22P05": "22—Data Exception",
	"2200L": "22—Data Exception",
	"2200M": "22—Data Exception",
	"2200N": "22—Data Exception",
	"2200S": "22—Data Exception",
	"2200T": "22—Data Exception",
	"22030": "22—Data Exception",
	"22031": "22—Data Exception",
	"22032": "22—Data Exception",
	"22033": "22—Data Exception",
	"22034": "22—Data Exception",
	"22035": "22—Data Exception",
	"22036": "22—Data Exception",
	"22037": "22—Data Exception",
	"22038": "22—Data Exception",
	"22039": "22—Data Exception",
	"2203A": "22—Data Exception",
	"2203B": "22—Data Exception",
	"2203C": "22—Data Exception",
	"2203D": "22—Data Exception",
	"2203E": "22—Data Exception",
	"2203F": "22—Data Exception",
	"2203G": "22—Data Exception",
	"23000": "23—Integrity Constraint Violation",
	"23001": "23—Integrity Constraint Violation",
	"23502": "23—Integrity Constraint Violation",
	"23503": "23—Integrity Constraint Violation",
	"23505": "23—Integrity Constraint Violation",
	"23514": "23—Integrity Constraint Violation",
	"23P01": "23—Integrity Constraint Violation",
	"24000": "24—Invalid Cursor State",
	"25000": "25—Invalid Transaction State",
	"25001": "25—Invalid Transaction State",
	"25002": "25—Invalid Transaction State",
	"25008": "25—Invalid Transaction State",
	"25003": "25—Invalid Transaction State",
	"25004": "25—Invalid Transaction State",
	"25005": "25—Invalid Transaction State",
	"25006": "25—Invalid Transaction State",
	"25007": "25—Invalid Transaction State",
	"25P01": "25—Invalid Transaction State",
	"25P02": "25—Invalid Transaction State",
	"25P03": "25—Invalid Transaction State",
	"26000": "26—Invalid SQL Statement Name",
	"27000": "27—Triggered Data Change Violation",
	"28000": "28—Invalid Authorization Specification",
	"28P01": "28—Invalid Authorization Specification",
	"2B000": "2B—Dependent Privilege Descriptors Still Exist",
	"2BP01": "2B—Dependent Privilege Descriptors Still Exist",
	"2D000": "2D—Invalid Transaction Termination",
	"2F000": "2F—SQL Routine Exception",
	"2F005": "2F—SQL Routine Exception",
	"2F002": "2F—SQL Routine Exception",
	"2F003": "2F—SQL Routine Exception",
	"2F004": "2F—SQL Routine Exception",
	"34000": "34—Invalid Cursor Name",
	"38000": "38—External Routine Exception",
	"38001": "38—External Routine Exception",
	"38002": "38—External Routine Exception",
	"38003": "38—External Routine Exception",
	"38004": "38—External Routine Exception",
	"39000": "39—External Routine Invocation Exception",
	"39001": "39—External Routine Invocation Exception",
	"39004": "39—External Routine Invocation Exception",
	"39P01": "39—External Routine Invocation Exception",
	"39P02": "39—External Routine Invocation Exception",
	"39P03": "39—External Routine Invocation Exception",
	"3B000": "3B—Savepoint Exception",
	"3B001": "3B—Savepoint Exception",
	"3D000": "3D—Invalid Catalog Name",
	"3F000": "3F—Invalid Schema Name",
	"40000": "40—Transaction Rollback",
	"40002": "40—Transaction Rollback",
	"40001": "40—Transaction Rollback",
	"40003": "40—Transaction Rollback",
	"40P01": "40—Transaction Rollback",
	"42000": "42—Syntax Error or Access Rule Violation",
	"42601": "42—Syntax Error or Access Rule Violation",
	"42501": "42—Syntax Error or Access Rule Violation",
	"42846": "42—Syntax Error or Access Rule Violation",
	"42803": "42—Syntax Error or Access Rule Violation",
	"42P20": "42—Syntax Error or Access Rule Violation",
	"42P19": "42—Syntax Error or Access Rule Violation",
	"42830": "42—Syntax Error or Access Rule Violation",
	"42602": "42—Syntax Error or Access Rule Violation",
	"42622": "42—Syntax Error or Access Rule Violation",
	"42939": "42—Syntax Error or Access Rule Violation",
	"42804": "42—Syntax Error or Access Rule Violation",
	"42P18": "42—Syntax Error or Access Rule Violation",
	"42P21": "42—Syntax Error or Access Rule Violation",
	"42P22": "42—Syntax Error or Access Rule Violation",
	"42809": "42—Syntax Error or Access Rule Violation",
	"428C9": "42—Syntax Error or Access Rule Violation",
	"42703": "42—Syntax Error or Access Rule Violation",
	"42883": "42—Syntax Error or Access Rule Violation",
	"42P01": "42—Syntax Error or Access Rule Violation",
	"42P02": "42—Syntax Error or Access Rule Violation",
	"42704": "42—Syntax Error or Access Rule Violation",
	"42701": "42—Syntax Error or Access Rule Violation",
	"42P03": "42—Syntax Error or Access Rule Violation",
	"42P04": "42—Syntax Error or Access Rule Violation",
	"42723": "42—Syntax Error or Access Rule Violation",
	"42P05": "42—Syntax Error or Access Rule Violation",
	"42P06": "42—Syntax Error or Access Rule Violation",
	"42P07": "42—Syntax Error or Access Rule Violation",
	"42712": "42—Syntax Error or Access Rule Violation",
	"42710": "42—Syntax Error or Access Rule Violation",
	"42702": "42—Syntax Error or Access Rule Violation",
	"42725": "42—Syntax Error or Access Rule Violation",
	"42P08": "42—Syntax Error or Access Rule Violation",
	"42P09": "42—Syntax Error or Access Rule Violation",
	"42P10": "42—Syntax Error or Access Rule Violation",
	"42611": "42—Syntax Error or Access Rule Violation",
	"42P11": "42—Syntax Error or Access Rule Violation",
	"42P12": "42—Syntax Error or Access Rule Violation",
	"42P13": "42—Syntax Error or Access Rule Violation",
	"42P14": "42—Syntax Error or Access Rule Violation",
	"42P15": "42—Syntax Error or Access Rule Violation",
	"42P16": "42—Syntax Error or Access Rule Violation",
	"42P17": "42—Syntax Error or Access Rule Violation",
	"44000": "44—WITH CHECK OPTION Violation",
	"53000": "53—Insufficient Resources",
	"53100": "53—Insufficient Resources",
	"53200": "53—Insufficient Resources",
	"53300": "53—Insufficient Resources",
	"53400": "53—Insufficient Resources",
	"54000": "54—Program Limit Exceeded",
	"54001": "54—Program Limit Exceeded",
	"54011": "54—Program Limit Exceeded",
	"54023": "54—Program Limit Exceeded",
	"55000": "55—Object Not In Prerequisite State",
	"55006": "55—Object Not In Prerequisite State",
	"55P02": "55—Object Not In Prerequisite State",
	"55P03": "55—Object Not In Prerequisite State",
	"55P04": "55—Object Not In Prerequisite State",
	"57000": "57—Operator Intervention",
	"57014": "57—Operator Intervention",
	"57P01": "57—Operator Intervention",
	"57P02": "57—Operator Intervention",
	"57P03": "57—Operator Intervention",
	"57P04": "57—Operator Intervention",
	"57P05": "57—Operator Intervention",
	"58000": "58—System Error (errors external to PostgreSQL itself)",
	"58030": "58—System Error (errors external to PostgreSQL itself)",
	"58P01": "58—System Error (errors external to PostgreSQL itself)",
	"58P02": "58—System Error (errors external to PostgreSQL itself)",
	"72000": "72—Snapshot Failure",
	"F0000": "F0—Configuration File Error",
	"F0001": "F0—Configuration File Error",
	"HV000": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV005": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV002": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV010": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV021": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV024": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV007": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV008": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV004": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV006": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV091": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00B": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00C": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00D": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV090": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00A": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV009": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV014": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV001": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00P": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00J": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00K": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00Q": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00R": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00L": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00M": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"HV00N": "HV—Foreign Data Wrapper Error (SQL/MED)",
	"P0000": "P0—PL/pgSQL Error",
	"P0001": "P0—PL/pgSQL Error",
	"P0002": "P0—PL/pgSQL Error",
	"P0003": "P0—PL/pgSQL Error",
	"P0004": "P0—PL/pgSQL Error",
}
