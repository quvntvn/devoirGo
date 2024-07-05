package main

type Product struct {
	id          int
	name        string
	description string
	price       float64
}

var products []Product
var nextID = 1

/*
Devoir GO
Échéance :7 juillet 2024 23:59
Instructions
Énoncé :

Développer une application CLI (Command Line Interface) en GO de gestion de produits comportant le menu suivant :
Ajouter un produit.
Afficher la liste des produits
Modifier un produit
Supprimer un produit
Exporter les informations produits dans un fichier Excel (en .xlsx)
Lancer un serveur Http avec une page web
Se connecter à une VM en ssh
Se connecter à un serveur FTP
(Bonus) Lancer l'interface graphique (qui comportera toutes les fonctionnalités du menu)
Quitter
Indications : La table Product contiendra les champs id, name, description, price.

Bonus : Ajouter une option pour lancer une interface graphique fonctionnelle (web via un serveur http réalisé en Go ou logiciel via la bibliothèque fyne par exemple)

lien de la bibliothèque fyne : fyne.io/fyne/v2

Attention : Les produits doivent être stockés dans une base MySql ou SqLite
Toute personne utilisant de manière abusive de l'IA verra sa note divisée par 2
Un projet comprenant trop de similitudes avec un autre, verra sa note divisée de manière récursive.
*/
