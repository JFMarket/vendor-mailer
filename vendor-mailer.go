package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"github.com/jfmarket/report-cacher/download"
	"github.com/mattbaird/gochimp"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

// Define program flags.
var (
	site         = flag.String("site", "https://jonesboroughfarmersmkt.shopkeepapp.com", "The address of the ShopKeep site reports will be retrieved from.")
	email        = flag.String("email", "", "The email used to login. (Required)")
	password     = flag.String("password", "", "The password used to login. (Required)")
	vendorEmails = flag.String("vendorEmails", "", "A CSV file containing vendor names and their email address. (Required)")
	key          = flag.String("key", "", "API Key from https://mandrillapp.com/settings/index (Required)")
	fromEmail    = flag.String("fromEmail", "", "The email address that vendors will see in the From field. (Required)")
	fromName     = flag.String("fromName", "", "The name vendors will see associated with the fromEmail. (Required)")
)

const tableTmpl = `<table>
  <tr>
    <th colspan="2">{{ .Name }}</th>
  <tr>
    <th>Item</th>
    <th>Quantity on Hand</th>
  </tr>
  {{range .Items }}
  <tr>
    <td>{{ .Name }}</td>
    <td align="right">{{ .Quantity }}</td>
  </tr>
  {{end}}
</table>`

// vendors is a slice of pointers to vendor to simplify csv parsing
type vendors []*vendor

// Pull in vendors from file p. This file should be a CSV with
// a header row that will be ignored and the following rows should be
// name, email.
func (v *vendors) getEmailsFromFile(p string) error {
	csvFile, err := os.Open(p)
	if err != nil {
		return errors.New("Could not open file " + p + " for reading: " + err.Error())
	}
	defer csvFile.Close()

	c := csv.NewReader(csvFile)

	lines, err := c.ReadAll()
	if err != nil {
		return errors.New("Could not parse CSV: " + err.Error())
	}

	// Start at 1 to skip the header row
	for i := 1; i < len(lines); i++ {
		for _, aVendor := range *v {
			if aVendor.Name == lines[i][0] {
				// This works because aVendor is a pointer to a vendor struct
				aVendor.Email = lines[i][1]
			}
		}
	}

	return nil
}

type vendor struct {
	Name  string
	Items []item
	Email string
}

type item struct {
	Name     string
	Quantity float64
}

func main() {
	// Parse and verify required options are set.
	flag.Parse()

	if *email == "" {
		log.Fatalln("An email is required. -email='x@yz.com'")
	}

	if *password == "" {
		log.Fatalln("A password is required. -password=mypassword")
	}

	if *vendorEmails == "" {
		log.Fatalln("The path to the vendor emails CSV file is required. -vendorEmails='vendoremails.csv'")
	}

	if *key == "" {
		log.Fatalln("An API Key from mandrillapp.com is required. -key='ad12410192ajkkea_G'")
	}

	if *fromEmail == "" {
		log.Fatalln("An email address specifying who reports are from is required. -fromEmail='john.doe@gmail.com'")
	}

	if *fromName == "" {
		log.Fatalln("A name specifying who reports are from is required. -fromName='John Doe'")
	}

	// Prepare temporary download directory.
	downloadDir, err := ioutil.TempDir("", "vendor-mailer")
	if err != nil {
		log.Fatalln("Failed to create temporary download directory: " + err.Error())
	}
	defer removeDir(downloadDir)

	// Download the stock items report
	stockItemsPath, err := downloadStockItemsReport(downloadDir)
	if err != nil {
		log.Fatalln(err)
	}

	v, err := stockCsvToVendors(stockItemsPath)
	// v, err := stockCsvToVendors("files/jonesboroughfarmersmkt_stock_items.csv") // remove this line when done testing.
	if err != nil {
		log.Fatalln(err)
	}

	err = v.getEmailsFromFile(*vendorEmails)
	if err != nil {
		log.Fatalln(err)
	}

	err = emailVendors(v)
	if err != nil {
		log.Fatalln(err)
	}
}

// emailVendors sends an inventory email to each vendor.
// This should probably be defined on vendors itself
func emailVendors(v vendors) error {
	tmpl, err := template.New("inventoryTable").Parse(tableTmpl)
	if err != nil {
		return err
	}

	// NewMandrill never actually returns an error.
	mandrill, _ := gochimp.NewMandrill(*key)
	// Check connection to mandrill
	_, err = mandrill.Ping()
	if err != nil {
		return errors.New("Failed to initialize mandrill. Bad Key?")
	}

	// Convenience function for sending vendor emails
	send := func(to string, name string, body string) ([]gochimp.SendResponse, error) {
		message := gochimp.Message{}
		message.AddRecipients(gochimp.Recipient{to, name})
		message.FromEmail = *fromEmail
		message.FromName = *fromName
		message.Subject = "Jonesborough Farmers Market Inventory Report for " + name
		message.Html = body

		return mandrill.MessageSend(message, false)
	}

	// email each vendor
	for _, aVendor := range v {
		// Skip vendors that don't have emails
		if aVendor.Email == "" {
			log.Println("No email found for " + aVendor.Name + ". Skipping...")
			continue
		}

		// Generate a table of items and quantities
		// for the vendor
		var t bytes.Buffer

		err = tmpl.Execute(&t, aVendor)
		if err != nil {
			return err
		}

		// Send the email
		_, err = send(aVendor.Email, aVendor.Name, t.String())
		if err != nil {
			// Perhaps just log the error instead of dying?
			return err
		}
		log.Println("Email sent to " + aVendor.Name + " (" + aVendor.Email + ")")
	}

	return nil
}

// stockCsvToVendors takes a path to the stock items report csv file
// and returns a vendors value and a nil error on success.
func stockCsvToVendors(p string) (vendors, error) {
	v := vendors{}

	csvFile, err := os.Open(p)
	if err != nil {
		return nil, errors.New("Could not open file " + p + " for reading: " + err.Error())
	}
	defer csvFile.Close()

	c := csv.NewReader(csvFile)

	lines, err := c.ReadAll()
	if err != nil {
		return nil, errors.New("Could not parse CSV: " + err.Error())
	}

	// Start at 1 to skip the header row
	for i := 1; i < len(lines); i++ {
		vendorName := lines[i][16]
		itemName := lines[i][1]
		itemQuantity, err := strconv.ParseFloat(lines[i][12], 64)
		if err != nil {
			log.Println("Failed parsing item quantity on " + itemName)
			continue
		}

		// inserted boolean allows us to create the vendor if it does not exist
		inserted := false
		for _, aVendor := range v {
			if aVendor.Name == vendorName {
				// This works because aVendor is a pointer to a vendor struct
				aVendor.Items = append(aVendor.Items, item{itemName, itemQuantity})
				inserted = true
			}
		}

		if !inserted {
			// Create a vendor and add its pointer to the slice
			v = append(v, &vendor{vendorName, []item{item{itemName, itemQuantity}}, ""})
		}
	}

	return v, nil
}

// Download the stock items report to the given directory.
// Returns the path to the report as a string.
// Error is non-nil if something goes wrong.
func downloadStockItemsReport(dir string) (string, error) {
	// downloader should be passed into the function if more
	// reports are necessary in the future.
	downloader, err := download.New(*site, *email, *password)
	if err != nil {
		return "", errors.New("Failed to initialize downloader: " + err.Error())
	}

	stockItemsPath := path.Join(dir, "stock_items.csv")

	err = downloader.GetStockItemsReport(stockItemsPath)
	if err != nil {
		return "", errors.New("Failed to get Stock Items report: " + err.Error())
	}

	return stockItemsPath, nil
}

// Remove the directory and all of its contents
// at the given path p.
func removeDir(p string) {
	err := os.RemoveAll(p)
	if err != nil {
		log.Println("Could not remove directory '" + p + "' :" + err.Error())
	}
}
