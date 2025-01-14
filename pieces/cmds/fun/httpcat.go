package fun

import (
	"math/rand"

	"github.com/opolobot/Opolo/ocl"
	"github.com/opolobot/Opolo/ocl/args"
	"github.com/opolobot/Opolo/pieces/parsers"
)

var statusCodes = []string{
	"100", "101", "102",
	"200", "201", "202", "203", "204", "205", "206", "207",
	"300", "301", "302", "303", "304", "305", "306", "307",
	"400", "401", "402", "403", "404", "405", "406", "408", "409",
	"410", "411", "412", "413", "414", "415", "416", "417", "418",
	"420", "421", "422", "423", "424", "425", "426", "429",
	"431", "444", "450", "451", "499",
	"500", "501", "502", "503", "504", "505", "506", "507", "508", "509",
	"510", "511", "599",
}

func init() {
	cmd := ocl.New()
	cmd.Name("httpcat")
	cmd.Aliases("cat", "http")
	cmd.Description("Grab a httpcat :3")
	cmd.Args(args.New("[code]", &parsers.String{}))
	cmd.Use(httpcat)

	Category.Add(cmd)
}

func httpcat(ctx *ocl.Context, next ocl.Next) {
	if len(ctx.RawArgs) > 0 {
		arg := ctx.Args["code"].(string)
		if stringInSlice(arg, statusCodes) {
			ctx.Send(httpCatURL(arg))
		} else if code := itemFromAInB(ctx.RawArgs, statusCodes); code != "" {
			ctx.Send(httpCatURL(code))
		} else if arg == "itjk" {
			// big floppa
			ctx.Send("https://piapiac.org/trash/floppa.webm")
		} else if arg == "zorbyte" {
			// SPEEN
			ctx.Send("https://piapiac.org/trash/SPEEN.webm")
		} else {
			ctx.Send(httpCatURL("404"))
		}
	} else {
		code := randomFromSlice(statusCodes)
		ctx.Send(httpCatURL(code))
	}
}

func httpCatURL(code string) string {
	return "https://http.cat/" + code
}

func randomFromSlice(slice []string) string {
	i := rand.Intn(len(slice))
	return slice[i]
}

func stringInSlice(str string, slice []string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func itemFromAInB(a []string, b []string) string {
	for _, item := range b {
		if stringInSlice(item, a) {
			return item
		}
	}
	return ""
}
