package main

import (
	"fmt"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run boot/manifests/main.go <service host, e.g. http://localhost:8000> [<org>]\nIf org is not specified an app is created for a user")
		os.Exit(1)
	}

	host := os.Args[1]
	orgName := os.Args[2]

	if _, err := url.Parse(host); err != nil {
		fmt.Println("host must be a valid URL, e.g. http://localhost:8000")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var baseURL string
	if orgName == "" {
		baseURL = "https://github.com/settings/apps/new"
	} else {
		baseURL = fmt.Sprintf("https://github.com/organizations/%s/settings/apps/new", orgName)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		os.Exit(1)
	}

	q := u.Query()
	q.Set("name", "Treenq")
	q.Set("description", "Treenq Platform-as-a-Service for Kubernetes")
	q.Set("url", "https://github.com/treenq/treenq")
	q.Add("callback_urls[]", host+"/authCallback")
	q.Set("webhook_active", "true")
	q.Set("webhook_url", host+"/githubWebhook")
	q.Set("public", "false")
	q.Set("request_oauth_on_install", "false")
	q.Set("setup_on_update", "true")

	// permissions
	q.Set("contents", "read")
	q.Set("metadata", "read")
	q.Set("pull_requests", "read")

	// webhook events
	q.Add("events[]", "push")
	q.Add("events[]", "pull_request")
	q.Add("events[]", "installation")
	q.Add("events[]", "installation_repositories")

	u.RawQuery = q.Encode()
	registrationURL := u.String()

	envFileName := "out.env"
	envFile, err := os.Create(envFileName)
	if err != nil {
		fmt.Printf("Error creating env file: %v\n", err)
		os.Exit(1)
	}
	defer envFile.Close()

	envContent := fmt.Sprintf(`# GitHub App Configuration for %s
DOCKER_REGISTRY=localhost:5005
DB_DSN=postgres://postgres@localhost:5432/tq?sslmode=disable
MIGRATIONS_DIR="./migrations"
HTTP_PORT=8000
AUTH_PRIVATE_KEY=LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpRZ0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1N3d2dna29BZ0VBQW9JQ0FRQy9CdEFzcDJCUUJDQ3EKVmpyZFRNRVRtNktYTTNuQzhsRGZLcCtPNUNzb3R1S2JwL2xsMzBSdTQ1czlqOS9mcEpnb1hxZ1BsK0ZZMHVrbApxU3U1NDVjRG9ZZVQvY1hhT093MVI3VEpKWVJkbHhFazZFeTBkL1Y0QXR2Q3BLdURndFNoNlMrZkR6cStpUUVXCnRncW1uLy9zUHFBb2JKVzVFNXBhdUZZb1NQeUlDSEo1K0xvUEJzZlQzWmsyYmxZdjJYOTlyWEI5dHZlT2pVeXAKbVBwaWFZdEZNUzFSZmlyOEwxVzc0RlZ1L0JEd0tpNFJjV1dOYjlqWHBVenJtWW5LS3Nqa3R3YTdubEtiOG1BYQpteHArOFppSXlCbUdLQkMzSnBPb0NIQUVUc1pvK3J2c1E5amVqVWREbTVyODgrUkc1d2c0U3VTVnB4ZDl2U2gxCi92RXA1RnVqK3NlSWVnbGV0NEZLYTR3UUNaRWR5RjA0a1FuSis1T0tEK2tLK201c1RYOENpaEdpWWxTSWtmZ08KVVphdzR3bW55NzRXRUtCb2ZpNk00WURTTkIrWTFYdmZkNzZDUUZ1eVJqNCtjckpLMHRUMkxmZmVjT3hXYXFRZApFVDJHNkp6bWRXRmxxOHc1d1hURlN2dFpuLzlrdlA4OHZ2R29Mdi9Sb3ZLZHFrcTdwYzMrS2hIdFNSa3RZYUx4CkpyZ3hqL1FUWEV4dlY5K1VFTEE1SW54TWx2Z3VBdUlMS3ZtQWtITFhzN1lDb09sQWw5SjNOcDhLTnZ6QXhCWnEKQlZ3cnI2bEJ4VEd3dUtzVkVodGpOTlM3Y0xPZnE4VUltR2NkQUxmcmovOWptWUp5aWZkMXR5em1pZ0hmYS80eQpXM2FDNEZyU25YNGVuT3I1d0Z6QWU0RFRHSmRNOVFJREFRQUJBb0lDQUNSMkMzUUtlb0tyVndUTU1xNGN3Vm1aCndqM3o4RkM2YVo4L0JuZUNxNDQ0NGlYdHVOZXQ5dVRuZ3JFTWJpSEV5OW9ndnhsQzF4dGFIbkEyeXdiUEh2cnQKY1BCWlp0TlJQQnlyVkNGMGpNQkVYbHhhRHBIL3Q1V3hqZnFuN2FqTFp3U0VlcmQwYzdUOGQzMjdQMnN3Yk92Tgp4RW9TOTd6OHRlQ05BTHp0dVczSmx2Z2E5b0I1dmRoOW1vbmVJNHM4aTI0VWxiMFpHRVZZU1FLeUZWQ0ZiclRGClB5NG4wOUtRd2w5NjRUT3UzeGpJSkVUbURRa2pDUk9ZRklkL3BlSTFxbVNsWHA2ZkNJdUZocnh0VUhCY2NzM2cKeTZaa3JWUVlBK1QxdjFQREtYSE9NZUpWek9ZbHR5MTdYT2pkTUh3c0N0Q2IwdWVSbmhVODdxVWZCUzhHekVxNgpTOTlJZzZ0UW1GREQzeWlVblhQdG5maEpxRUFmN2hXalJpVXdNeDBTdWkxVG5zQ056Y3o5NlFMRS91c01iTml3CkdDQW50ZWpEQXhEVFFNRURwN21WUGlJY25HSjhYOEJMT0YvWVBWSE0vdHdpN0VhODBIbVA4dEtVWTNiYWlmYkYKVis5WElpQnBPcmFaWG5Va1hWMng0Nzl4K250ZEVPc2RxQ1FNbmloN3VIWkZ6MEdWelU5SkZ5bW9uSUFNNWdNRwpaRUZlVnlUVUxwSGJoaDhsdUdMOW5hem1lVU5FbkpYWlFSRDkyTDUzY0dnWUR0RTVIejR5Rm1lTnBGbjI2TmpVCjdUd2xMN3M5eGdGaVdBZzhraUN1K1MxRWltakc5WlRHOWs3VnJCaHBFRjEwNmNIdEt5SXlmQmIzMjE1MFdqYlQKTnU3aDFLUzN3M2wzeGt1ZUxOeGRBb0lCQVFENGZPaWpZRVc1aGw0MFh3b3dNUlhzSnlGUHhBWUJuTzUxZDQ3VQp6T1ZDa1kzNUt5VjRkQ1RoMGRxYXJvSjFLb0ExYjJySnpydkxEVzhoQlJTcVpZZXE3YUtvZGFDM2FmZHlBQ1BLClB1NU1DT1JDd2o2ajJSVWV6VHNVOXhuWXdHb0JDTkVDeHptbGp3ZW9VM3pBZ0JqdDNOTVpkUnVmTXlwNlM0SUkKQ01yU1RmYlRaVGphT3lEQ2tCZlE4d0pEREdoVEtac29ZT01lUE05VWMzN3hiYmcva0ZtOUM4SzI1OFlwUFE5QQp5eGRueHpreFp4VHI2ZUh2REpRa25WYkpHVS9jVE14VVEzcis3VnJIQ0V0WnBRR3RWN2VTQkpqVkJodDdlTWlCCjVGemNnYklYdnkzdE1wWkxQb2VVQ09JSGJMaDA1bmt1NlhzUlZJZGd3UzJXbU9uVEFvSUJBUURFelRPTWhBOGEKRzJUTEJMenFZWTN5ZXNZekhyam1xUWhvNXIzUUh2UWE5NXJoa3FOZ2ZHWjZMcWtrMzR0TEpQQmFSQnFYTDFCaAo3VDEzYmxTRnVKRjJzVFNXZ1A2cGRXNDZhWGY4emhxRkJndWVzNjArc2F0QnVMOUFaeSthemI2c2ZQNndldFFLCk83MUQ2N3hSQWZjelVzUG5ZdHV3anZ6R0VQc1VHZ0xNRW9Ub2F0d25xNzV2bGNVNDZUWG5ERTF3bFF6bkxqMGcKb2YzT3hZVXVzQ20yeDk2U25wdVNzZ3VoTmd0eHc2ZzJPaDRHRHpuL0U1OER4TkEvdjBUN3llTXQ3QTJ6OUtrQgpGVjU2YXNsRDdoSjY1azFlRUl2ZzZEdy9tbnRVb0R0WjByVHNTckFadW54SFIyZG9rSmpaNWhOL290YTNBK0RKCkxzdWNOM3VxeXFrWEFvSUJBUURWZVFtckFaUWs5RWlPR1cvNVF0SWdsT1ZMVDQ0UmFLNTdnQURXMUVmSXpwNzMKaHBla1NiTTl4VGxXVmNHQndzZHVJSS9QVzZsOW9jYnN5UjZkM0tlV3NweGd1TjBmZFF1OWhsMDQ3S052OHR1VwpkcVcxd1ZNaHFSS3V5aklNUWhGUFhqR2hmMklJMXoydTREcDJiaFg4a3c5UCtZbUhWVCtTM0xlVEMycWpEWk9VCnZJc0JBSGIrYnlmbXZENGZOOU9RVGxnYmNsRHJzell5eVI0dmJ5RXdpbVJ0d21LL0c1TGo3cTdoN1Jmb2NnQ3kKYm1wTTJocmRjU0w3NmFlYkVBSEpzcmgvVTVHZzNHeDJQS3Z2RVpERlNHeE9KMkRjOFdnK3hOOE1xQkVXNXQwSApCWmtCQThxV1RkdlAwMm5MRXgzTVlBdVB2OG1ZYzlQeHpVUEs4d2M5QW9JQkFCbjF5NWY4ajVWdENhV3lNVTFsCm92amFjeXlwSDlEbGVVT0ZOSUt3b3Bpd0V0RXdxN1o5a25NSmxxeFRoS2RiN3d5cE1TekNSQU0yN1VYRTJ3ZHMKcWx1UzBwSUw0QXZ3ZnFMYjZNVURWd0kzSXU5RFdsWUx6OEJ5bzEvV2ZMUVo4YzRGQ3YreXBDZlphNFQ3SXJNNAo1Q2YwQWYyU3o5SUJlcHlSL3R6TzlaRi8yK0pndmp2SmJ1eC9RQzNhclk2VjA5MUcvQlcreHJkNFJ1ZXdySG5WCktSdUFULzdkUno1Wm1Da2kzTzJiMXFPWWxQOU5vT1BoN2Jic2psL3FWaW8wbm5BZlZFdHB1YVYzOFNlSzBKUFMKWXNPdGY0VXAvNW1pYU5nbkE1L05KeWNaSVY2T0Y1NjlOOG1iUkt2SmJ6QkhKa2xPN0szbHFCQkJ5UUFKWFpuUQo4KzhDZ2dFQWVyUkdtUXRGWjFleXdoWGxKN3pBaWV2eEZOb2EwbXhwemlyeFI4NzZIVWhIVUtETVV6cFE3aEJJCmNCSlRJUzRRUFVnVlBZREc2MmFwKys3d2treDFBUFRSa0U4VE9rT2xvOXZQRGpVUlAyREVMTVRibVBZM1pCYzEKWXhSdFlYSFZ3Vmh3MTI0R1FBTzZqaWJCdWZaajBaWnNLMFRGYXMwVUpyUXRzWjR5SGp3R2pOVXZ1RU13SVYyOQpneFhhZFpZbnVzQlRaVGJUZkt1TllWRUo1UGlKU0l6SlNNR0dDNEU0VkRHZ3EwWkVQNGVqQzRRTTlmZDZMYnVHCnI4OGJtUnNiUTZ2RmlpbmJ3c2xMZnF6bk1raTdVU1lrYzZ0MFhWRitEQ0E4Nlk3SlBUektGbktDTkNxRHQ0Q0IKOURxNnoxVkVsYXhGTDVIcno4R0xoNzUzN0tLT1pnPT0KLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=
AUTH_PUBLIC_KEY=LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUF2d2JRTEtkZ1VBUWdxbFk2M1V6QgpFNXVpbHpONXd2SlEzeXFmanVRcktMYmltNmY1WmQ5RWJ1T2JQWS9mMzZTWUtGNm9ENWZoV05McEpha3J1ZU9YCkE2R0hrLzNGMmpqc05VZTB5U1dFWFpjUkpPaE10SGYxZUFMYndxU3JnNExVb2Vrdm53ODZ2b2tCRnJZS3BwLy8KN0Q2Z0tHeVZ1Uk9hV3JoV0tFajhpQWh5ZWZpNkR3YkgwOTJaTm01V0w5bC9mYTF3ZmJiM2pvMU1xWmo2WW1tTApSVEV0VVg0cS9DOVZ1K0JWYnZ3UThDb3VFWEZsalcvWTE2Vk02NW1KeWlySTVMY0d1NTVTbS9KZ0dwc2FmdkdZCmlNZ1poaWdRdHlhVHFBaHdCRTdHYVBxNzdFUFkzbzFIUTV1YS9QUGtSdWNJT0Vya2xhY1hmYjBvZGY3eEtlUmIKby9ySGlIb0pYcmVCU211TUVBbVJIY2hkT0pFSnlmdVRpZy9wQ3ZwdWJFMS9Bb29Sb21KVWlKSDREbEdXc09NSgpwOHUrRmhDZ2FINHVqT0dBMGpRZm1OVjczM2UrZ2tCYnNrWStQbkt5U3RMVTlpMzMzbkRzVm1xa0hSRTlodWljCjVuVmhaYXZNT2NGMHhVcjdXWi8vWkx6L1BMN3hxQzcvMGFMeW5hcEt1NlhOL2lvUjdVa1pMV0dpOFNhNE1ZLzAKRTF4TWIxZmZsQkN3T1NKOFRKYjRMZ0xpQ3lyNWdKQnkxN08yQXFEcFFKZlNkemFmQ2piOHdNUVdhZ1ZjSzYrcApRY1V4c0xpckZSSWJZelRVdTNDem42dkZDSmhuSFFDMzY0Ly9ZNW1DY29uM2RiY3M1b29CMzJ2K01sdDJndUJhCjBwMStIcHpxK2NCY3dIdUEweGlYVFBVQ0F3RUFBUT09Ci0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQo=
AUTH_TTL=60m
AUTH_REDIRECT_URL=http://localhost:9000
KUBE_CONFIG=k3s_data/k3s/k3s.yaml
BUILDKIT_HOST=tcp://localhost:1234
BUILDKIT_TLS_CA=./buildkit/certs/ca.crt
HOST=localhost
REGISTRY_CERT=./registry/certs/ca.crt
REGISTRY_TLS_VERIFY=true
REGISTRY_AUTH_TYPE=basic
REGISTRY_AUTH_USERNAME=testuser
REGISTRY_AUTH_PASSWORD=testpassword
REGISTRY_AUTH_TOKEN=
CORS_ALLOW_ORIGIN='http://localhost:9000'
IS_PROD=false

#### delete this line and fill the config below ####### 
GITHUB_CLIENT_ID=
GITHUB_SECRET=
GITHUB_PRIVATE_KEY=
GITHUB_WEBHOOK_SECRET=
GITHUB_WEBHOOK_SECRET_ENABLE=false
GITHUB_REDIRECT_URL=%s/authCallback
`, orgName, host, host)

	_, err = envFile.WriteString(envContent)
	if err != nil {
		fmt.Printf("Error writing to env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("GitHub App Registration for %s\n", orgName)
	fmt.Printf("===================================\n\n")
	fmt.Printf("1. Open this URL in your browser:\n")
	fmt.Printf("   %s\n\n", registrationURL)
	fmt.Printf("2. Review the pre-filled configuration and create the app\n\n")
	fmt.Printf("3. After creating the app, go to the app settings page and:\n")
	fmt.Printf("   - Copy the Client ID\n")
	fmt.Printf("   - Generate and copy a Client Secret\n")
	fmt.Printf("   - Generate and copy the Private Key (PEM format)\n")
	fmt.Printf("   - Copy or generate the Webhook Secret\n\n")
	fmt.Printf("4. Fill in the values in the generated file: %s\n\n", envFileName)
	fmt.Printf("5. Check the generated app slug and update APP_GITHUB_APP_NAME in web/.env\n\n")
}
