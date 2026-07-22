package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"

	"pixivhono/client"
	"pixivhono/config"
	"pixivhono/lib"
)

type ctxKey string

const baseURLCtxKey ctxKey = "baseURL"

// GraphQLHandler manages GraphQL schema and execution.
type GraphQLHandler struct {
	cfg    *config.Config
	schema graphql.Schema
}

// NewGraphQLHandler initializes the GraphQL schema and returns the handler.
func NewGraphQLHandler(cfg *config.Config) *GraphQLHandler {
	h := &GraphQLHandler{cfg: cfg}
	h.initSchema()
	return h
}

func (h *GraphQLHandler) initSchema() {
	// ── JSON scalar for dynamic fields (e.g. _resolved) ──
	jsonScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:         "JSON",
		Description:  "Arbitrary JSON value",
		Serialize:    func(v interface{}) interface{} { return v },
		ParseValue:   func(v interface{}) interface{} { return v },
		ParseLiteral: func(v ast.Value) interface{} { return v.GetValue() },
	})

	// ── ImageURLs type ──────────────────────────────────
	imageURLsType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ImageURLs",
		Fields: graphql.Fields{
			"square_medium": &graphql.Field{Type: graphql.String},
			"medium":        &graphql.Field{Type: graphql.String},
			"large":         &graphql.Field{Type: graphql.String},
			"original":      &graphql.Field{Type: graphql.String},
			"_resolved":     &graphql.Field{Type: jsonScalar},
		},
	})

	// ── Tag type ────────────────────────────────────────
	tagType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Tag",
		Fields: graphql.Fields{
			"name":            &graphql.Field{Type: graphql.String},
			"translated_name": &graphql.Field{Type: graphql.String},
		},
	})

	// ── Illust type ─────────────────────────────────────
	illustType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Illust",
		Fields: graphql.Fields{
			"id":              &graphql.Field{Type: graphql.Int},
			"title":           &graphql.Field{Type: graphql.String},
			"caption":         &graphql.Field{Type: graphql.String},
			"type":            &graphql.Field{Type: graphql.String},
			"image_urls":      &graphql.Field{Type: imageURLsType},
			"width":           &graphql.Field{Type: graphql.Int},
			"height":          &graphql.Field{Type: graphql.Int},
			"page_count":      &graphql.Field{Type: graphql.Int},
			"sanity_level":    &graphql.Field{Type: graphql.Int},
			"x_restrict":      &graphql.Field{Type: graphql.Int},
			"total_view":      &graphql.Field{Type: graphql.Int},
			"total_bookmarks": &graphql.Field{Type: graphql.Int},
			"create_date":     &graphql.Field{Type: graphql.String},
			"tags":            &graphql.Field{Type: graphql.NewList(tagType)},
			"user":            &graphql.Field{Type: graphql.String},
		},
	})

	// ── Query type ──────────────────────────────────────
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"illust": &graphql.Field{
				Type: illustType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: h.resolveIllust,
			},
			"search": &graphql.Field{
				Type: graphql.NewList(illustType),
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"page":  &graphql.ArgumentConfig{Type: graphql.Int},
					"limit": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: h.resolveSearch,
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		panic(fmt.Sprintf("Failed to init GraphQL schema: %v", err))
	}
	h.schema = schema
}

func (h *GraphQLHandler) resolveIllust(p graphql.ResolveParams) (interface{}, error) {
	id, ok := p.Args["id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid id argument")
	}

	refreshToken := h.cfg.PixivRefreshToken
	if refreshToken == "" {
		return nil, fmt.Errorf("missing env: PIXIV_REFRESH_TOKEN")
	}

	accessToken, err := client.GetPixivAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	resp, err := client.PixivGet("/v1/illust/detail", map[string]string{"illust_id": strconv.Itoa(id)}, accessToken)
	if err != nil {
		return nil, fmt.Errorf("pixiv api: %w", err)
	}

	baseURL, _ := p.Context.Value(baseURLCtxKey).(string)
	enriched := lib.EnrichArtworkResponseWithResolvedUrls(resp, baseURL)

	m, ok := enriched.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	illust, ok := m["illust"]
	if !ok {
		return nil, fmt.Errorf("no illust in response")
	}
	return illust, nil
}

func (h *GraphQLHandler) resolveSearch(p graphql.ResolveParams) (interface{}, error) {
	query, ok := p.Args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	page := 1
	if pv, ok2 := p.Args["page"].(int); ok2 && pv > 0 {
		page = pv
	}
	limit := 30
	if lv, ok2 := p.Args["limit"].(int); ok2 && lv > 0 {
		limit = lv
	}
	// ponytail: cap at 100, bump if clients need more
	if limit > 100 {
		limit = 100
	}

	refreshToken := h.cfg.PixivRefreshToken
	if refreshToken == "" {
		return nil, fmt.Errorf("missing env: PIXIV_REFRESH_TOKEN")
	}

	accessToken, err := client.GetPixivAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	const perPage = 30
	params := map[string]string{
		"word":                           query,
		"search_target":                  "partial_match_for_tags",
		"sort":                           "date_desc",
		"filter":                         "for_ios",
		"merge_plain_keyword_results":    "true",
		"include_translated_tag_results": "true",
		"offset":                         strconv.Itoa((page - 1) * perPage),
	}
	resp, err := client.PixivGet("/v1/search/illust", params, accessToken)
	if err != nil {
		return nil, fmt.Errorf("pixiv api: %w", err)
	}

	baseURL, _ := p.Context.Value(baseURLCtxKey).(string)
	enriched := lib.EnrichSearchResponseWithResolvedUrls(resp, baseURL)

	m, ok := enriched.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	illusts, ok := m["illusts"].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}
	if len(illusts) > limit {
		illusts = illusts[:limit]
	}
	return illusts, nil
}

// GraphQLRequest defines the JSON payload for GraphQL queries.
type GraphQLRequest struct {
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
}

// Handle processes GraphQL requests (POST /api/graphql).
func (h *GraphQLHandler) Handle(c *fiber.Ctx) error {
	var req GraphQLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx := context.WithValue(c.Context(), baseURLCtxKey, c.BaseURL())
	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	if len(result.Errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(result)
	}
	return c.JSON(result)
}

// Playground serves the GraphiQL interactive UI (GET /graphql).
func (h *GraphQLHandler) Playground(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html")
	return c.SendString(graphiqlHTML)
}

const graphiqlHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>PixivHono GraphiQL</title>
  <style>
    body { height:100vh; margin:0; width:100%; display:flex; flex-direction:column; overflow:hidden; font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif; background:#fafafa; }
    #settings { padding:16px 24px; background:#fff; border-bottom:1px solid #eaeaea; display:flex; flex-wrap:wrap; gap:12px; align-items:center; box-shadow:0 1px 3px rgba(0,0,0,0.04); }
    #settings input { padding:8px 12px; border:1px solid #e1e4e8; border-radius:6px; font-size:14px; outline:none; transition:all .2s ease; background:#f6f8fa; }
    #settings input:focus { border-color:#0366d6; box-shadow:0 0 0 3px rgba(3,102,214,0.3); background:#fff; }
    #settings button { padding:8px 16px; background:#2ea44f; color:#fff; border:1px solid rgba(27,31,35,0.15); border-radius:6px; cursor:pointer; font-size:14px; font-weight:500; transition:all .2s ease; }
    #settings button:hover { background:#2c974b; }
    #graphiql { flex:1; }
    @media (max-width:600px) { #settings { flex-direction:column; align-items:stretch; } #settings input,#settings button { width:100%!important; box-sizing:border-box; } }
  </style>
  <script src="https://unpkg.com/react@17/umd/react.production.min.js"></script>
  <script src="https://unpkg.com/react-dom@17/umd/react-dom.production.min.js"></script>
  <link rel="stylesheet" href="https://unpkg.com/graphiql/graphiql.min.css" />
</head>
<body>
  <div id="settings">
    <input type="text" id="api-url" placeholder="GraphQL URL" value="/api/graphql" style="width:250px" />
    <input type="text" id="api-key" placeholder="API Key" value="" />
    <button onclick="updateFetcher()">Apply</button>
  </div>
  <div id="graphiql">Loading...</div>
  <script src="https://unpkg.com/graphiql/graphiql.min.js"></script>
  <script>
    function render() {
      var u=document.getElementById('api-url').value||'/api/graphql',k=document.getElementById('api-key').value,h={};
      if(k) h['Authorization']='Bearer '+k;
      ReactDOM.render(React.createElement(GraphiQL,{fetcher:GraphiQL.createFetcher({url:u,headers:h})}),document.getElementById('graphiql'));
    }
    function updateFetcher(){ReactDOM.unmountComponentAtNode(document.getElementById('graphiql'));render()}
    render();
  </script>
</body>
</html>`
