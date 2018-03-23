package activedir

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mavricknz/ldap"
)

func GetGroupsForUser(userID string) ([]string, error) {
	log.Printf("--CHECK 1")
	groups := []string{}
	log.Printf("--CHECK 2")
	conn := ldap.NewLDAPConnection(
		"cad3.byu.edu",
		389)
	log.Printf("--CHECK 3")
	err := conn.Connect()
	log.Printf("--CHECK 4")
	if err != nil {
		log.Printf("--CHECK 4a")
		panic(err)
	}
	log.Printf("--CHECK 5")
	defer conn.Close()
	log.Printf("--CHECK 6")
	username := os.Getenv("LDAP_USERNAME")
	log.Printf("--CHECK 7")
	password := os.Getenv("LDAP_PASSWORD")
	log.Printf("--CHECK 8")
	err = conn.Bind(username, password)
	log.Printf("--CHECK 9")
	if err != nil {
		log.Printf("--CHECK 9a")
		panic(err)
	}
	log.Printf("--CHECK 10")
	search := ldap.NewSearchRequest(
		"ou=people,dc=byu,dc=local",
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,
		0,
		false,
		fmt.Sprintf("(Name=%s)", userID),
		[]string{"Name", "MemberOf"},
		nil,
	)
	log.Printf("--CHECK 11")
	res, err := conn.Search(search)
	log.Printf("--CHECK 12")
	if err != nil {
		log.Printf("--CHECK 12a")
		panic(err)
	}
	log.Printf("--CHECK 13")
	//verify name
	for i := 0; i < len(res.Entries); i++ {
		log.Printf("--CHECK 14")
		name := res.Entries[i].GetAttributeValue("Name")
		if name != userID {
			log.Printf("--CHECK 15")
			continue
		}
		log.Printf("--CHECK 16")

		groupsTemp := res.Entries[0].GetAttributeValues("MemberOf")
		log.Printf("--CHECK 17")
		groups = translateGroups(groupsTemp)
		log.Printf("--CHECK 18")
	}

	return groups, nil
}

func translateGroups(groups []string) []string {
	toReturn := []string{}

	for _, entry := range groups {
		AD := strings.Split(entry, ",")
		toReturn = append(toReturn, strings.TrimPrefix(AD[0], "CN="))
	}
	return toReturn
}
