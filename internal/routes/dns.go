package routes

import (
	"fmt"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	"github.com/sirupsen/logrus"
)

var list = []string{
	"bl.spamcop.net",
	"b.barracudacentral.org",
	"ydqnrpeoxvkopd2y7klgywch4y.sbl-xbl.dq.spamhaus.net",
}

func reverse(ip net.IP) string {
	octets := strings.Split(ip.String(), ".")
	for i, j := 0, len(octets)-1; i < j; i, j = i+1, j-1 {
		octets[i], octets[j] = octets[j], octets[i]
	}
	return strings.Join(octets, ".")
}

type lookupResult struct {
  Exists  bool `json:"exists"`
  Reasons []string `json:"reasons"`
}

func exists(ip, server string) (*lookupResult, error) {
	addr := fmt.Sprintf("%v.%v", ip, server)
	res, err := net.LookupHost(addr)
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return &lookupResult{
				Exists: false,
			}, nil
		}
		return nil, err
	}
	result := &lookupResult{}
	for _, v := range res {
		if strings.HasPrefix(v, "127.0.0.") {
			result.Exists = true
		}
	}
	if result.Exists {
		records, _ := net.LookupTXT(addr)
		for _, v := range records {
			result.Reasons = append(result.Reasons, v)
		}
	}
	return result, nil
}

func testDnsRoute(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Address string `json:"ip"`
		}
		err := c.BodyParser(&body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		if body.Address == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		addr := net.ParseIP(body.Address)
		if addr == nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Address",
			})
		}
    logrus.Info("yo")
		reversedIP := reverse(addr)
    var checks = make(map[string]*lookupResult)
		for _, v := range list {
			result, err := exists(reversedIP, v)
			if err != nil {
				logrus.Error("dns check err: ", err)
				continue
			}
			checks[v] = result
		}
    logrus.Info("CHECKS: ", checks)
		return c.JSON(checks)
	}
}
