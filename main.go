package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	//"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jlaffaye/ftp"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/ssh"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var w fyne.Window

func main() {
	fmt.Println("\n----------------------- Menu principal -----------------------\n")
	fmt.Println("Utiliser num pad (0-9) :")
	fmt.Println("\n1 - Ajouter un produit.")
	fmt.Println("2 - Afficher la liste des produits")
	fmt.Println("3 - Modifier un produit")
	fmt.Println("4 - Supprimer un produit")
	fmt.Println("5 - Exporter les informations produits dans un fichier Excel (en .xlsx)")
	fmt.Println("6 - Lancer un serveur Http avec une page web")
	fmt.Println("7 - Se connecter à une VM en ssh")
	fmt.Println("8 - Se connecter à un serveur FTP")
	fmt.Println("9 - Lancer l'interface graphique")
	fmt.Println("0 - Quitter")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nTapez un numéro pour choisir une option : ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var err error
	db, err = sql.Open("sqlite", "./products.db")
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture de la base de données :", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL
	)
`)
	if err != nil {
		fmt.Println("Erreur lors de la création de la table :", err)
		return
	}

	switch input {
	case "1":
		if err := addProductCLI(); err != nil {
			fmt.Println("Erreur lors de l'ajout du produit :", err)
		}
	case "2":
		listProduct()
	case "3":
		updateProductCLI()
	case "4":
		deleteProduct()
	case "5":
		exportProduct()
	case "6":
		startServer()
	case "7":
		connectSSH()
	case "8":
		connectFTP()
	case "9":
		startGUI()
	case "0":
		os.Exit(0)
	default:
		fmt.Println("Choix invalide veuillez saisir un numéro valide (0-9)")
		main()
	}
}

func addProductCLI() error {
	fmt.Println("\n----------------------- Ajouter un produit -----------------------\n")
	fmt.Println("Veuillez saisir les informations du produit :\n\nNom : ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Println("Description : ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Println("Prix : ")
	price, _ := reader.ReadString('\n')
	price = strings.TrimSpace(price)

	err := addProduct(db, name, description, price)
	if err != nil {
		return fmt.Errorf("\n\nerreur lors de l'ajout du produit : %v", err)
	}

	fmt.Println("\n\n" + name + " produit ajouté avec succès !\n\n")
	fmt.Println("Voulez-vous ajouter un autre produit ? (O/N)")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "O" || choice == "o" {
		return addProductCLI()
	} else {
		main()
	}
	return nil
}

func addProduct(db *sql.DB, name, description, price string) error {
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return fmt.Errorf("erreur lors de la conversion du prix : %v", err)
	}

	_, err = db.Exec("INSERT INTO products (name, description, price) VALUES (?, ?, ?)", name, description, priceFloat)
	if err != nil {
		return fmt.Errorf("erreur lors de l'insertion du produit : %v", err)
	}
	return nil
}

func listProduct() {
	fmt.Println("\n----------------------- Liste des produits -----------------------\n")

	rows, err := db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		fmt.Println("Erreur lors de la récupération des produits :", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, description string
		var price float64
		err := rows.Scan(&id, &name, &description, &price)
		if err != nil {
			fmt.Println("Erreur lors de la lecture des données :", err)
			return
		}

		fmt.Printf("ID: %d\nNom: %s\nDescription: %s\nPrix: %.2f€\n\n", id, name, description, price)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Erreur lors de l'itération des produits :", err)
		return
	}

	fmt.Println("\nRetourner au menu principal (entrée)")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	main()
}

func updateProductCLI() {
	fmt.Println("\n----------------------- Modifier un produit -----------------------\n")

	rows, err := db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		fmt.Println("Erreur lors de la récupération des produits :", err)
		return
	}
	defer rows.Close()

	var products []string
	fmt.Println("ID\tNom\tDescription\tPrix")
	for rows.Next() {
		var id int
		var name, description string
		var price float64
		err := rows.Scan(&id, &name, &description, &price)
		if err != nil {
			fmt.Println("Erreur lors de la lecture des données :", err)
			return
		}
		fmt.Printf("%d\t%s\t%s\t%.2f€\n", id, name, description, price)
		products = append(products, name)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Erreur lors de l'itération des produits :", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEntrez l'ID du produit à modifier : ")
	idInput, _ := reader.ReadString('\n')
	idInput = strings.TrimSpace(idInput)
	productID, err := strconv.Atoi(idInput)
	if err != nil {
		fmt.Println("ID invalide :", err)
		return
	}

	var name, description string
	var price float64
	err = db.QueryRow("SELECT name, description, price FROM products WHERE id = ?", productID).Scan(&name, &description, &price)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Produit non trouvé")
		} else {
			fmt.Println("Erreur lors de la récupération du produit :", err)
		}
		return
	}

	fmt.Printf("Nom actuel (%s) : ", name)
	newName, _ := reader.ReadString('\n')
	newName = strings.TrimSpace(newName)
	if newName == "" {
		newName = name
	}

	fmt.Printf("Description actuelle (%s) : ", description)
	newDescription, _ := reader.ReadString('\n')
	newDescription = strings.TrimSpace(newDescription)
	if newDescription == "" {
		newDescription = description
	}

	fmt.Printf("Prix actuel (%.2f) : ", price)
	newPrice, _ := reader.ReadString('\n')
	newPrice = strings.TrimSpace(newPrice)
	var newPriceFloat float64
	if newPrice == "" {
		newPriceFloat = price
	} else {
		newPriceFloat, err = strconv.ParseFloat(newPrice, 64)
		if err != nil {
			fmt.Println("Prix invalide :", err)
			return
		}
	}

	_, err = db.Exec("UPDATE products SET name = ?, description = ?, price = ? WHERE id = ?", newName, newDescription, newPriceFloat, productID)
	if err != nil {
		fmt.Println("Erreur lors de la mise à jour du produit :", err)
		return
	}

	fmt.Println("\nProduit mis à jour avec succès")

	fmt.Println("\nVoulez-vous modifier un autre produit ? (O/N)")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "O" || choice == "o" {
		updateProductCLI()
	} else {
		main()
	}
}

func deleteProduct() {
	fmt.Println("\n----------------------- Supprimer un produit -----------------------\n")

	rows, err := db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		fmt.Println("Erreur lors de la récupération des produits :", err)
		return
	}
	defer rows.Close()

	fmt.Println("ID\tNom\tDescription\tPrix")
	for rows.Next() {
		var id int
		var name, description string
		var price float64
		err := rows.Scan(&id, &name, &description, &price)
		if err != nil {
			fmt.Println("Erreur lors de la lecture des données :", err)
			return
		}
		fmt.Printf("%d\t%s\t%s\t%.2f€\n", id, name, description, price)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Erreur lors de l'itération des produits :", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEntrez l'ID du produit à supprimer : ")
	idInput, _ := reader.ReadString('\n')
	idInput = strings.TrimSpace(idInput)
	productID, err := strconv.Atoi(idInput)
	if err != nil {
		fmt.Println("ID invalide :", err)
		return
	}

	var name string
	err = db.QueryRow("SELECT name FROM products WHERE id = ?", productID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Produit non trouvé")
		} else {
			fmt.Println("Erreur lors de la récupération du produit :", err)
		}
		return
	}

	fmt.Printf("Êtes-vous sûr de vouloir supprimer le produit '%s' ? (O/N) : ", name)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "O" || choice == "o" {
		_, err := db.Exec("DELETE FROM products WHERE id = ?", productID)
		if err != nil {
			fmt.Println("Erreur lors de la suppression du produit :", err)
			return
		}
		fmt.Println("Produit supprimé avec succès")
	}

	fmt.Println("Voulez-vous supprimer un autre produit ? (O/N)")
	choice, _ = reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "O" || choice == "o" {
		deleteProduct()
	} else {
		main()
	}
}

func exportProduct() {
	fmt.Println("\n----------------------- Exporter les informations produits dans un fichier Excel (en .xlsx) -----------------------\n")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	headers := []string{"ID", "Nom", "Description", "Prix"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue("Sheet1", cell, header); err != nil {
			fmt.Println("Erreur lors de l'ajout des en-têtes :", err)
			return
		}
	}

	rows, err := db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		fmt.Println("Erreur lors de la récupération des produits :", err)
		return
	}
	defer rows.Close()

	row := 2
	for rows.Next() {
		var id int
		var name, description string
		var price float64
		if err := rows.Scan(&id, &name, &description, &price); err != nil {
			fmt.Println("Erreur lors de la lecture des données :", err)
			return
		}

		if err := f.SetCellValue("Sheet1", "A"+strconv.Itoa(row), id); err != nil {
			fmt.Println("Erreur lors de l'ajout de la valeur ID :", err)
			return
		}
		if err := f.SetCellValue("Sheet1", "B"+strconv.Itoa(row), name); err != nil {
			fmt.Println("Erreur lors de l'ajout de la valeur Nom :", err)
			return
		}
		if err := f.SetCellValue("Sheet1", "C"+strconv.Itoa(row), description); err != nil {
			fmt.Println("Erreur lors de l'ajout de la valeur Description :", err)
			return
		}
		if err := f.SetCellValue("Sheet1", "D"+strconv.Itoa(row), price); err != nil {
			fmt.Println("Erreur lors de l'ajout de la valeur Prix :", err)
			return
		}

		row++
	}

	if err := f.SaveAs("Produits.xlsx"); err != nil {
		fmt.Println("Erreur lors de la sauvegarde du fichier Excel :", err)
		return
	}

	fmt.Println("Fichier Excel 'Produits.xlsx' créé avec succès")

	fmt.Println("\nRetourner au menu principal (entrée)")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	main()
}

func startServer() {
	fmt.Println("\n----------------------- Lancer un serveur Http avec une page web -----------------------\n")
	http.HandleFunc("/", homeHandler)
	fmt.Println("Serveur en écoute sur http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func addProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		description := r.FormValue("description")
		price := r.FormValue("price")

		err := addProduct(db, name, description, price)
		if err != nil {
			http.Error(w, "Erreur lors de l'ajout du produit", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func connectSSH() {
	fmt.Println("\n----------------------- Se connecter à une VM en SSH -----------------------\n")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Entrez l'adresse IP de la VM : ")
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Entrez le nom d'utilisateur : ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Entrez le mot de passe : ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("\nEntrez la commande à exécuter : ")
	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		fmt.Println("Erreur lors de la connexion SSH :", err)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		fmt.Println("Erreur lors de la création de la session SSH :", err)
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution de la commande :", err)
		return
	}

	fmt.Printf("Résultat de la commande :\n%s\n", output)

	fmt.Println("\nRetourner au menu principal (entrée)")
	reader.ReadString('\n')
	main()
}

func connectFTP() {
	fmt.Println("\n----------------------- Se connecter à un serveur FTP -----------------------\n")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Entrez l'adresse IP ou le nom d'hôte du serveur FTP : ")
	serverAddr, _ := reader.ReadString('\n')
	serverAddr = strings.TrimSpace(serverAddr)

	fmt.Print("Entrez le nom d'utilisateur FTP : ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Entrez le mot de passe FTP : ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	ftpClient, err := ftp.Connect(serverAddr)
	if err != nil {
		fmt.Println("Erreur lors de la connexion au serveur FTP :", err)
		return
	}
	defer ftpClient.Quit()

	err = ftpClient.Login(username, password)
	if err != nil {
		fmt.Println("Erreur lors de l'authentification FTP :", err)
		return
	}
	fmt.Println("\nConnecté au serveur FTP avec succès")

	files, err := ftpClient.List(".")
	if err != nil {
		fmt.Println("Erreur lors de la récupération de la liste des fichiers :", err)
		return
	}

	fmt.Println("\nListe des fichiers dans le répertoire courant :")
	for _, file := range files {
		fmt.Println(file.Name)
	}

	fmt.Println("\nRetourner au menu principal (appuyez sur Entrée)")
	reader.ReadString('\n')
	main()
}

func startGUI() {
	fmt.Println("\n----------------------- Lancer l'interface graphique -----------------------\n")
	fmt.Println("Choisissez le type d'interface graphique:\n")
	fmt.Println("1 - Interface Web")
	fmt.Println("2 - Interface Desktop (Fyne)\n")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Tapez un numéro pour choisir une option : ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		startWebGUI()
	case "2":
		startDesktopGUI()
	default:
		fmt.Println("Choix invalide. Retour au menu principal.")
		main()
	}
}

func startWebGUI() {
	fmt.Println("Lancer l'interface graphique web")
	startServer()
}

func startDesktopGUI() {
	fmt.Println("Lancer l'interface graphique desktop (Fyne)")
	a := app.New()
	w = a.NewWindow("Gestion de Produits - Menu Principal")

	addProductButton := widget.NewButton("Ajouter un produit", func() {
		openAddProductWindow(a)
	})
	listProductButton := widget.NewButton("Afficher la liste des produits", func() {
		openListProductWindow(a)
	})
	updateProductButton := widget.NewButton("Modifier un produit", func() {
		openUpdateProductWindow(a)
	})
	deleteProductButton := widget.NewButton("Supprimer un produit", func() {
		openDeleteProductWindow(a)
	})
	exportProductButton := widget.NewButton("Exporter les informations produits dans un fichier Excel (en .xlsx)", func() {
		exportProduct()
	})
	startServerButton := widget.NewButton("Lancer un serveur Http avec une page web", func() {
		startServer()
	})
	connectSSHButton := widget.NewButton("Se connecter à une VM en ssh", func() {
		connectSSH()
	})
	connectFTPButton := widget.NewButton("Se connecter à un serveur FTP", func() {
		connectFTP()
	})

	menu := container.NewVBox(
		addProductButton,
		listProductButton,
		updateProductButton,
		deleteProductButton,
		exportProductButton,
		startServerButton,
		connectSSHButton,
		connectFTPButton,
		widget.NewButton("Quitter", func() {
			a.Quit()
		}),
	)

	w.SetContent(menu)
	w.Resize(fyne.NewSize(400, 400))
	w.ShowAndRun()
}

func openAddProductWindow(a fyne.App) {
	w := a.NewWindow("Ajouter un produit")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom du produit")
	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Description du produit")
	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Prix du produit")

	addButton := widget.NewButton("Ajouter le produit", func() {
		name := nameEntry.Text
		description := descriptionEntry.Text
		price := priceEntry.Text
		err := addProduct(db, name, description, price)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Erreur",
				Content: fmt.Sprintf("Erreur lors de l'ajout du produit : %v", err),
			})
		} else {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Succès",
				Content: fmt.Sprintf("Produit %s ajouté avec succès !", name),
			})
			nameEntry.SetText("")
			descriptionEntry.SetText("")
			priceEntry.SetText("")
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Ajouter un produit"),
		nameEntry,
		descriptionEntry,
		priceEntry,
		addButton,
		widget.NewButton("Retour au menu principal", func() {
			w.Close()
		}),
	)

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 200))
	w.Show()
}

func openListProductWindow(a fyne.App) {
	w := a.NewWindow("Liste des produits")

	data := binding.NewStringList()

	productList := widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			obj.(*widget.Label).Bind(item.(binding.String))
		},
	)

	productList.Resize(fyne.NewSize(800, 600))

	refreshProductList := func() {
		var products []string
		rows, err := db.Query("SELECT name, price FROM products")
		if err != nil {
			fmt.Println("Erreur lors de la récupération des produits :", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			var price float64
			err := rows.Scan(&name, &price)
			if err != nil {
				fmt.Println("Erreur lors de la lecture des données :", err)
				return
			}
			products = append(products, fmt.Sprintf("%s - %.2f€", name, price))
		}
		data.Set(products)
	}

	scroll := container.NewScroll(productList)
	scroll.SetMinSize(fyne.NewSize(800, 600))

	refreshButton := widget.NewButton("Rafraîchir", refreshProductList)

	w.SetContent(container.NewVBox(
		widget.NewLabel("Liste des produits"),
		productList,
		refreshButton,
		widget.NewButton("Retour au menu principal", func() {
			w.Close()
		}),
	))

	refreshProductList()
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

func openUpdateProductWindow(a fyne.App) {
	w := a.NewWindow("Modifier un produit")

	productData := binding.NewStringList()
	productMap := make(map[string]int)

	loadProducts := func() {
		rows, err := db.Query("SELECT id, name FROM products")
		if err != nil {
			fmt.Println("Erreur lors de la récupération des produits :", err)
			return
		}
		defer rows.Close()
		var products []string
		for rows.Next() {
			var id int
			var name string
			err := rows.Scan(&id, &name)
			if err != nil {
				fmt.Println("Erreur lors de la lecture des données :", err)
				return
			}
			products = append(products, name)
			productMap[name] = id
		}
		productData.Set(products)
	}

	productList := widget.NewListWithData(
		productData,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			obj.(*widget.Label).Bind(item.(binding.String))
		},
	)

	nameEntry := widget.NewEntry()
	descriptionEntry := widget.NewEntry()
	priceEntry := widget.NewEntry()

	var selectedProductID int

	productList.OnSelected = func(id widget.ListItemID) {
		productName, _ := productData.GetValue(id)
		selectedProductID = productMap[productName]

		var name, description string
		var price float64

		err := db.QueryRow("SELECT name, description, price FROM products WHERE id = ?", selectedProductID).Scan(&name, &description, &price)
		if err != nil {
			fmt.Println("Erreur lors de la récupération des détails du produit :", err)
			return
		}

		nameEntry.SetText(name)
		descriptionEntry.SetText(description)
		priceEntry.SetText(fmt.Sprintf("%.2f", price))
	}

	updateButton := widget.NewButton("Modifier le produit", func() {
		name := nameEntry.Text
		description := descriptionEntry.Text
		price := priceEntry.Text

		priceFloat, err := strconv.ParseFloat(price, 64)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Erreur",
				Content: fmt.Sprintf("Erreur lors de la conversion du prix : %v", err),
			})
			return
		}

		_, err = db.Exec("UPDATE products SET name = ?, description = ?, price = ? WHERE id = ?", name, description, priceFloat, selectedProductID)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Erreur",
				Content: fmt.Sprintf("Erreur lors de la mise à jour du produit : %v", err),
			})
		} else {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Succès",
				Content: fmt.Sprintf("Produit %s modifié avec succès !", name),
			})
			loadProducts()
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Modifier un produit"),
		productList,
		nameEntry,
		descriptionEntry,
		priceEntry,
		updateButton,
		widget.NewButton("Retour au menu principal", func() {
			w.Close()
		}),
	)

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 400))
	w.Show()

	loadProducts()
}

func openDeleteProductWindow(a fyne.App) {

	w := a.NewWindow("Supprimer un produit")

	productData := binding.NewStringList()
	productMap := make(map[string]int)

	loadProducts := func() {
		rows, err := db.Query("SELECT id, name FROM products")
		if err != nil {
			fmt.Println("Erreur lors de la récupération des produits :", err)
			return
		}
		defer rows.Close()
		var products []string
		for rows.Next() {
			var id int
			var name string
			err := rows.Scan(&id, &name)
			if err != nil {
				fmt.Println("Erreur lors de la lecture des données :", err)
				return
			}
			products = append(products, name)
			productMap[name] = id
		}
		productData.Set(products)
	}

	productList := widget.NewListWithData(
		productData,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			obj.(*widget.Label).Bind(item.(binding.String))
		},
	)

	var selectedProductID int

	productList.OnSelected = func(id widget.ListItemID) {
		productName, _ := productData.GetValue(id)
		selectedProductID = productMap[productName]
	}

	deleteButton := widget.NewButton("Supprimer le produit", func() {
		confirmDialog := dialog.NewConfirm("Confirmation", fmt.Sprintf("Êtes-vous sûr de vouloir supprimer ce produit ?"), func(confirmed bool) {
			if confirmed {
				_, err := db.Exec("DELETE FROM products WHERE id = ?", selectedProductID)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Erreur",
						Content: fmt.Sprintf("Erreur lors de la suppression du produit : %v", err),
					})
				} else {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Succès",
						Content: "Produit supprimé avec succès !",
					})
					loadProducts()
				}
			}
		}, w)

		confirmDialog.SetDismissText("Annuler")
		confirmDialog.Show()
	})

	form := container.NewVBox(
		widget.NewLabel("Supprimer un produit"),
		productList,
		deleteButton,
		widget.NewButton("Retour au menu principal", func() {
			w.Close()
		}),
	)

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 400))
	w.Show()

	loadProducts()
}
