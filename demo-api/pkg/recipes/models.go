package recipes

// Represents a recipe
type Recipe struct {
	Name        string       `json:"name"`
	Ingredients []Ingredient `json:"ingredients"`
}

// Represents individual ingredients
type Ingredient struct {
	Name string `json:"name"`
}

// Take note of the json struct tag in the code above. As you're using a JSON payload in this REST API, you add the json struct tag to define the name of the field in JSON representation. This will be useful later to encode and decode JSON into the recipe struct.
