package expert

import (
	"fmt"
	"os"
)

// Deprecated: Remove in v2.0.
var legacyAliases = map[string]string{
	// Go
	"rob-pike":    "the-go-purist",
	"dave-cheney": "the-go-purist",
	"antirez":     "the-go-purist",

	// Ruby
	"matz":       "the-ruby-crafter",
	"sandi-metz": "the-ruby-crafter",
	"avdi-grimm": "the-ruby-crafter",

	// Rails
	"dhh":             "the-rails-monolith",
	"jorge-manrubia":  "the-rails-monolith",
	"rafael-franca":   "the-rails-monolith",
	"eileen-uchitelle": "the-activerecord-surgeon",
	"tenderlove":       "the-activerecord-surgeon",

	// Python
	"raymond-hettinger": "the-pythonista",
	"simon-willison":    "the-django-pragmatist",
	"carlton-gibson":    "the-django-pragmatist",

	// Swift
	"chris-lattner":    "the-protocol-thinker",
	"john-sundell":     "the-protocol-thinker",
	"soroush-khanlou": "the-protocol-thinker",

	// Kotlin/Android
	"jake-wharton":    "the-kotlin-architect",
	"roman-elizarov":  "the-kotlin-architect",
	"gabor-varadi":    "the-kotlin-architect",

	// Frontend
	"jeremy-keith":     "the-html-fundamentalist",
	"luke-wroblewski":  "the-html-fundamentalist",

	// JavaScript/React
	"dan-abramov": "the-react-philosopher",

	// Elixir
	"jose-valim":  "the-otp-alchemist",
	"sasa-juric":  "the-otp-alchemist",

	// Phoenix
	"jose-valim-phoenix":   "the-liveview-builder",
	"sophie-debenedetto":   "the-liveview-builder",
	"chris-mccord":         "the-liveview-builder",

	// TypeScript
	"anders-hejlsberg": "the-type-guardian",
	"matt-pocock":      "the-type-guardian",

	// Java
	"joshua-bloch":  "the-api-classicist",
	"martin-fowler": "the-api-classicist",

	// C#
	"anders-hejlsberg-csharp": "the-net-pragmatist",
	"nick-chapsas":            "the-net-pragmatist",

	// PHP
	"taylor-otwell": "the-laravel-artisan",
	"nuno-maduro":   "the-laravel-artisan",

	// Vue
	"evan-you":    "the-vue-reactivity-nerd",
	"anthony-fu":  "the-vue-reactivity-nerd",

	// Node
	"ryan-dahl":       "the-event-loop-guy",
	"matteo-collina":  "the-event-loop-guy",

	// Rust
	"steve-klabnik": "the-borrow-checker",

	// Clojure
	"rich-hickey":  "the-data-thinker",
	"alex-miller":  "the-data-thinker",

	// Next.js
	"guillermo-rauch": "the-edge-deployer",
	"lee-robinson":    "the-edge-deployer",

	// Django
	// simon-willison already mapped above to the-django-pragmatist

	// Svelte
	"rich-harris": "the-compiler-whisperer",
	"tan-li-hau":  "the-compiler-whisperer",

	// Flutter
	"eric-seidel":     "the-widget-composer",
	"remi-rousselet":  "the-widget-composer",

	// C++
	"bjarne-stroustrup": "the-zero-cost-abstracter",
	"john-carmack":      "the-zero-cost-abstracter",

	// Scala
	"martin-odersky": "the-type-theorist",
	"john-de-goes":   "the-type-theorist",

	// GraphQL
	"lee-byron":          "the-schema-purist",
	"marc-andre-giroux":  "the-schema-purist",

	// Writing
	"william-zinsser": "the-revision-hawk",
	"stephen-king":    "the-revision-hawk",

	// Business
	"paul-graham": "the-startup-realist",

	// Product
	"marty-cagan":    "the-product-skeptic",
	"teresa-torres":  "the-product-skeptic",
	"des-traynor":    "the-product-skeptic",

	// Design/UX
	"don-norman":    "the-usability-scientist",
	"jakob-nielsen": "the-usability-scientist",
	"julie-zhuo":    "the-usability-scientist",

	// Growth
	"brian-balfour":  "the-growth-mechanic",
	"april-dunford": "the-growth-mechanic",
	"sean-ellis":    "the-growth-mechanic",

	// Sales
	"mark-roberge": "the-revenue-engineer",
	"aaron-ross":   "the-revenue-engineer",

	// Finance
	"david-skok": "the-metrics-hawk",
	"brad-feld":  "the-metrics-hawk",

	// Leadership
	"andy-grove":           "the-eng-manager",
	"camille-fournier":     "the-eng-manager",
	"will-larson":          "the-eng-manager",

	// Operations
	"elad-gil":              "the-scale-operator",
	"ben-horowitz":          "the-scale-operator",
	"claire-hughes-johnson": "the-scale-operator",

	// Hiring
	"geoff-smart": "the-hiring-bar-raiser",
	"laszlo-bock": "the-hiring-bar-raiser",

	// Data
	"avinash-kaushik": "the-data-storyteller",
	"dj-patil":        "the-data-storyteller",

	// Security
	"bruce-schneier": "the-threat-modeler",

	// Legal
	"heather-meeker": "the-license-auditor",

	// Customer Success
	"lincoln-murphy": "the-retention-strategist",
	"nick-mehta":     "the-retention-strategist",

	// Bootstrapped
	"rob-walling":   "the-bootstrap-realist",
	"sahil-lavingia": "the-bootstrap-realist",
	"arvid-kahl":     "the-bootstrap-realist",

	// VC
	"reid-hoffman":    "the-venture-strategist",
	"keith-rabois":    "the-venture-strategist",
	"marc-andreessen": "the-venture-strategist",

	// General
	"kent-beck":    "the-tdd-advocate",
	"jason-fried":  "the-scope-cutter",
	"dieter-rams":  "the-design-minimalist",
	"gene-kim":     "the-flow-optimizer",
	"cal-newport":  "the-deep-worker",
	"taiichi-ohno": "the-flow-optimizer",
}

// LegacyAlias resolves a deprecated real-name ID to its composite replacement.
// Deprecated: Remove in v2.0.
func LegacyAlias(id string) (string, bool) {
	if newID, ok := legacyAliases[id]; ok {
		fmt.Fprintf(os.Stderr, "Warning: %q is deprecated, use %q instead\n", id, newID)
		return newID, true
	}
	return id, false
}
