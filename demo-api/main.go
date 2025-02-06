// based on https://www.jetbrains.com/guide/go/tutorials/rest_api_series/stdlib/

package main

import (
	"encoding/json"
	"net/http"
	"regexp"

	"joshuamkite/go-api-demo/pkg/recipes"

	"github.com/gosimple/slug"
)

func main() {
	// Create the Store and Recipe Handler
	store := recipes.NewMemStore()
	recipesHandler := NewRecipesHandler(store)

	// create a new request multiplexer
	// take incoming requessts and despatch them to matching handlers
	mux := http.NewServeMux()
	// Register the routes and handlers
	mux.Handle("/", &homeHandler{})
	mux.Handle("/recipes", recipesHandler)
	mux.Handle("/recipes/", recipesHandler)
	// run the server
	http.ListenAndServe(":8080", mux)
}

type RecipesHandler struct {
	store recipeStore
}

func (h *RecipesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && RecipeRe.MatchString(r.URL.Path):
		h.CreateRecipe(w, r)
		return
	case r.Method == http.MethodGet && RecipeRe.MatchString(r.URL.Path):
		h.ListRecipes(w, r)
		return
	case r.Method == http.MethodGet && RecipeReWithID.MatchString(r.URL.Path):
		h.GetRecipe(w, r)
	case r.Method == http.MethodPut && RecipeReWithID.MatchString(r.URL.Path):
		h.UpdateRecipe(w, r)
		return
	case r.Method == http.MethodDelete && RecipeReWithID.MatchString(r.URL.Path):
		h.DeleteRecipe(w, r)
		return
	default:
		return
	}
}

func NewRecipesHandler(s recipeStore) *RecipesHandler {
	return &RecipesHandler{
		store: s,
	}
}

type homeHandler struct{}

// define handler for homepage route
func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my home page"))
}

// create one function per action in your table and use a switch in ServeHTTP to route the request to the matching function. Each function signature will match a HandlerFunc type. That way, ServeHTTP can easily delegate the task of handling the request but propagate w http.ResponseWriter and r *http.Request to the next function.

func (h *RecipesHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Convert the name of the recipe into URL friendly string
	resourceID := slug.Make(recipe.Name)
	// Call the store to add the recipe
	if err := h.store.Add(resourceID, recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Set the status code to 200
	w.WriteHeader(http.StatusOK)
}

func (h *RecipesHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	resources, err := h.store.List()
	jsonBytes, err := json.Marshal(resources)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using a regex
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
	// Expect matches to be length >= 2 (full string + 1 matching group)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	// Retrieve recipe from the store
	recipe, err := h.store.Get(matches[1])
	if err != nil {
		// Special case of NotFound Error
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}

		// Every other error
		InternalServerErrorHandler(w, r)
		return
	}

	// Convert the struct into JSON payload
	jsonBytes, err := json.Marshal(recipe)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Write the results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	// Recipe object will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Update(matches[1], recipe); err != nil {
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *RecipesHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Remove(matches[1]); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader((http.StatusOK))
}

var (
	RecipeRe       = regexp.MustCompile(`^/recipes/*$`)
	RecipeReWithID = regexp.MustCompile(`^/recipes/([a-z0-9]+(?:-[a-z0-9]+)+)$`)
)

// An interface allows your server code to be totally detached from the storage solution (in memory, Redis, MySQL, etc.). You can store your data anywhere as long as you implement this interface.
type recipeStore interface {
	Add(name string, recipe recipes.Recipe) error
	Get(name string) (recipes.Recipe, error)
	Update(name string, recipe recipes.Recipe) error
	List() (map[string]recipes.Recipe, error)
	Remove(name string) error
}

// InternalServerErrorHandler and NotFoundHandler, which will be useful when dealing with errors:

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}
