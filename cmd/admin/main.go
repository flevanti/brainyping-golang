package main

import (
	"errors"
	"fmt"
	"os"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/utilities"

	"go.mongodb.org/mongo-driver/mongo/options"
)

const GOBACK string = ".."

func main() {
	initapp.InitApp("ADMIN")
	mainMenu()
}

func mainMenu() {

	for {
		var options [][]string
		options = append(options, []string{"createcol", "Create a missing collection"})
		options = append(options, []string{"trcol", "Truncate a collection and rebuild the indexes (all records lost!)"})
		options = append(options, []string{"dropcol", "Drop a collection (all records lost!)"})
		options = append(options, []string{"showcol", "Show the collections list"})
		options = append(options, []string{"showconfig", "Show the configuration settings"})
		options = append(options, []string{"m", "Show this menu"})
		options = append(options, []string{"q", "Quit"})
		utilities.PrintTable([]string{"CMD", "DESCRIPTION"}, options)

	internalLoop:
		for {
			option := utilities.ReadUserInputWithOptions("", []string{"createcol", "trcol", "dropcol", "showcol", "showconfig", "m", "h", "q"}, "")
			switch option {
			case "createcol":
				createCollectionMenu()
				break internalLoop
			case "trcol":
				truncateCollectionMenu()
				break internalLoop
			case "dropcol":
				dropCollectionMenu()
				break internalLoop
			case "showcol":
				showCollections()
				break
			case "showconfig":
				showConfig()
			case "q":
				os.Exit(0)
			case "m", "h":
				break internalLoop
			default:
				utilities.FailOnError(errors.New("something went wrong, command not found"))
			}
		}
	}
}

func showConfig() {
	var settsTable [][]string
	setts, err := initapp.GetSettings()
	utilities.FailOnError(err)
	for _, v := range setts {
		settsTable = append(settsTable, []string{v.Key, v.Value, v.Description})
	}

	utilities.PrintTable([]string{"KEY", "VALUE", "DESCRIPTION"}, settsTable)

}

func createCollectionMenu() {
	fmt.Println("List of current collections")
	showCollections()

	collection := utilities.ReadUserInput("Name of the new collection? ")
	if collection == GOBACK || collection == "" {
		return
	}

	utilities.FailOnError(dbhelper.CreateCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), collection, &options.CreateCollectionOptions{}))

}

func showCollections() {
	cl := dbhelper.CollectionsList(dbhelper.GetClient(), dbhelper.GetDatabaseName())
	utilities.PrintTableOneColumn("Collection", cl)
}

func dropCollectionMenu() {
	cl := dbhelper.CollectionsList(dbhelper.GetClient(), dbhelper.GetDatabaseName())
	utilities.PrintTableOneColumn("Collection", cl)
	collection := utilities.ReadUserInputWithOptions("Which collection you want to drop?", cl, GOBACK)
	if collection == GOBACK {
		return
	}
	if !utilities.ReadUserInputConfirm(fmt.Sprintf("Are you sure you want to drop [%s].[%s]?", dbhelper.GetDatabaseName(), collection)) {
		return
	}
	err := dbhelper.DeleteCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), collection)
	if err != nil {
		utilities.FailOnError(err)
	}
}

func truncateCollectionMenu() {
	cl := dbhelper.CollectionsList(dbhelper.GetClient(), dbhelper.GetDatabaseName())
	utilities.PrintTableOneColumn("Collection", cl)
	collection := utilities.ReadUserInputWithOptions("Which collection you want to truncate?", cl, GOBACK)
	if collection == GOBACK {
		return
	}
	if !utilities.ReadUserInputConfirm(fmt.Sprintf("Are you sure you want to truncate [%s].[%s]?", dbhelper.GetDatabaseName(), collection)) {
		return
	}
	err := dbhelper.TruncateCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), collection)
	if err != nil {
		utilities.FailOnError(err)
	}
}
