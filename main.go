package main

import (
	"fmt"
	"log"

	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Person struct {
	ID       *models.RecordID     `json:"id,omitempty"`
	Name     string               `json:"name"`
	Surname  string               `json:"surname"`
	Email    string               `json:"email"`
	Location models.GeometryPoint `json:"location"`
	Friends  []models.RecordID    `json:"friends,omitempty"`
}

type Post struct {
	ID      *models.RecordID `json:"id,omitempty"`
	Author  models.RecordID  `json:"author"`
	Content string           `json:"content"`
	Likes   int              `json:"likes"`
}

func main() {
	// Initialisation de la base de données
	db, err := surrealdb.New("ws://localhost:8000")
	if err != nil {
		log.Fatal("Erreur de connexion:", err)
	}
	defer db.Close()

	// Configuration du namespace et de la base
	if err = db.Use("socialapp", "main"); err != nil {
		log.Fatal("Erreur d'utilisation:", err)
	}

	// Authentification
	authData := &surrealdb.Auth{
		Username: "root",
		Password: "root",
	}

	token, err := db.SignIn(authData)
	if err != nil {
		log.Fatal("Erreur d'authentification:", err)
	}

	if err = db.Authenticate(token); err != nil {
		log.Fatal("Erreur de token:", err)
	}

	// Nettoyage à la fin
	defer func() {
		if err := db.Invalidate(); err != nil {
			log.Fatal("Erreur de déconnexion:", err)
		}
	}()

	// 1. Création d'utilisateurs avec géolocalisation
	fmt.Println("🚀 Création d'utilisateurs...")

	person1, err := surrealdb.Create[Person](db, models.Table("persons"), Person{
		Name:     "Emmanuel",
		Surname:  "Manou",
		Email:    "emmanuel@example.com",
		Location: models.NewGeometryPoint(-0.11, 22.00), // Bouaké, Côte d'Ivoire

	})
	if err != nil {
		log.Fatal("Erreur création person1:", err)
	}
	fmt.Printf("✅ Utilisateur créé: %+v\n", person1)

	person2, err := surrealdb.Create[Person](db, models.Table("persons"), Person{
		Name:     "Tilonon",
		Surname:  "Tilonon",
		Email:    "marie@example.com",
		Location: models.NewGeometryPoint(2.3522, 48.8566), // Paris, France

	})
	if err != nil {
		log.Fatal("Erreur création person2:", err)
	}
	fmt.Printf("✅ Utilisateur créé: %+v\n", person2)

	// 2. Création de relations d'amitié
	fmt.Println("\n🤝 Création de relations...")

	query := `RELATE $person1->friends->$person2 SET created_at = time::now();`
	_, err = surrealdb.Query[any](db, query, map[string]any{
		"person1": person1.ID,
		"person2": person2.ID,
	})
	if err != nil {
		log.Fatal("Erreur relation:", err)
	}
	fmt.Println("✅ Relation d'amitié créée")

	// 3. Création de posts
	fmt.Println("\n📝 Création de posts...")

	post1, err := surrealdb.Create[Post](db, models.Table("posts"), Post{
		Author:  *person1.ID,
		Content: "Salut depuis Bouaké ! SurrealDB est incroyable 🚀",
		Likes:   0,
	})
	if err != nil {
		log.Fatal("Erreur création post:", err)
	}
	fmt.Printf("✅ Post créé: %s\n", post1.Content)

	// 4. Requête complexe : posts des amis avec géolocalisation
	fmt.Println("\n🔍 Recherche des posts des amis...")

	complexQuery := `
		SELECT
			->friends->posts.* as friend_posts,
			->friends->location as friend_locations,
			geo::distance(location, {
				type: "Point",
				coordinates: [2.3522, 48.8566]
			}) as distance_to_paris
		FROM $person_id
		WHERE ->friends->posts IS NOT EMPTY;
`

	result, err := surrealdb.Query[any](db, complexQuery, map[string]any{
		"person_id": person2.ID,
	})
	if err != nil {
		log.Fatal("Erreur requête complexe:", err)
	}
	fmt.Printf("✅ Résultat requête: %+v\n", result)

	// 5. Mise à jour en temps réel (simulation)
	fmt.Println("\n⚡ Simulation temps réel...")

	// Écoute des changements (en arrière-plan)
	go func() {
		liveQuery := `LIVE SELECT * FROM posts WHERE author = $author_id;`
		_, err := surrealdb.Query[any](db, liveQuery, map[string]any{
			"author_id": person1.ID,
		})
		if err != nil {
			log.Printf("Erreur live query: %v", err)
		}
	}()

	// Mise à jour du nombre de likes
	updateQuery := `UPDATE $post_id SET likes = likes + 1;`
	_, err = surrealdb.Query[any](db, updateQuery, map[string]any{
		"post_id": post1.ID,
	})
	if err != nil {
		log.Fatal("Erreur mise à jour:", err)
	}
	fmt.Println("✅ Like ajouté en temps réel")

	// 6. Recherche géospatiale
	fmt.Println("\n🌍 Recherche géospatiale...")

	// version Go
	geoQuery := `
		SELECT *, geo::distance(location, {
			type: "Point",
			coordinates: [2.3522, 48.8566]
		}) as distance_km
		FROM persons
		WHERE geo::distance(location, {
			type: "Point",
			coordinates: [2.3522, 48.8566]
		}) < 10000000
		ORDER BY distance_km ASC;
`

	geoResult, err := surrealdb.Query[any](db, geoQuery, nil)
	if err != nil {
		log.Fatal("Erreur géospatiale:", err)
	}
	fmt.Printf("✅ Utilisateurs par distance de Paris: %+v\n", geoResult)

	fmt.Println("\n🎉 Démonstration terminée ! SurrealDB c'est magique ✨")
}
