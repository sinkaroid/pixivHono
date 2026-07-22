package lib

import (
	"net/url"
	"strings"

	"pixivhono/config"
)

func normalizeResolverBase(baseUrl string) string {
	forceHttpsEnabled := config.GlobalConfig.ForceHttpsPixivImgResolver
	if forceHttpsEnabled {
		if strings.HasPrefix(strings.ToLower(baseUrl), "http://") {
			return "https://" + baseUrl[7:]
		}
	}
	return baseUrl
}

func makeResolvedUrl(baseUrl, raw string) string {
	externalResolver := strings.TrimSpace(config.GlobalConfig.PixivImgResolver)
	if externalResolver != "" {
		resolverBaseRaw := externalResolver
		if !strings.HasPrefix(strings.ToLower(externalResolver), "http") {
			resolverBaseRaw = "https://" + externalResolver
		}
		resolverBase := normalizeResolverBase(resolverBaseRaw)
		// Remove trailing slashes
		for strings.HasSuffix(resolverBase, "/") {
			resolverBase = resolverBase[:len(resolverBase)-1]
		}
		u, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		queryPart := ""
		if u.RawQuery != "" {
			queryPart = "?" + u.RawQuery
		}
		return resolverBase + u.Path + queryPart
	}

	normalizedBase := normalizeResolverBase(baseUrl)
	return normalizedBase + "/pixiv/img_resolver?url=" + url.QueryEscape(raw)
}

func addResolvedImageUrls(imageUrls interface{}, baseUrl string) interface{} {
	m, ok := imageUrls.(map[string]interface{})
	if !ok {
		return imageUrls
	}

	resolved := make(map[string]interface{})
	for k, v := range m {
		if s, ok := v.(string); ok && strings.HasPrefix(s, "https://") {
			resolved[k] = makeResolvedUrl(baseUrl, s)
		}
	}

	next := make(map[string]interface{})
	for k, v := range m {
		next[k] = v
	}
	next["_resolved"] = resolved
	return next
}

func enrichIllust(illust interface{}, baseUrl string) interface{} {
	m, ok := illust.(map[string]interface{})
	if !ok {
		return illust
	}

	next := make(map[string]interface{})
	for k, v := range m {
		next[k] = v
	}

	if imageUrls, ok := next["image_urls"]; ok {
		next["image_urls"] = addResolvedImageUrls(imageUrls, baseUrl)
	}

	if metaSinglePage, ok := next["meta_single_page"].(map[string]interface{}); ok {
		nextMetaSingle := make(map[string]interface{})
		for k, v := range metaSinglePage {
			nextMetaSingle[k] = v
		}

		if origUrl, ok := nextMetaSingle["original_image_url"].(string); ok && strings.HasPrefix(origUrl, "https://") {
			nextMetaSingle["original_image_url_resolved"] = makeResolvedUrl(baseUrl, origUrl)
		}
		next["meta_single_page"] = nextMetaSingle
	}

	if metaPages, ok := next["meta_pages"].([]interface{}); ok {
		nextMetaPages := make([]interface{}, len(metaPages))
		for i, page := range metaPages {
			if pageMap, ok := page.(map[string]interface{}); ok {
				nextPageMap := make(map[string]interface{})
				for k, v := range pageMap {
					nextPageMap[k] = v
				}
				if imgUrls, ok := nextPageMap["image_urls"]; ok {
					nextPageMap["image_urls"] = addResolvedImageUrls(imgUrls, baseUrl)
				}
				nextMetaPages[i] = nextPageMap
			} else {
				nextMetaPages[i] = page
			}
		}
		next["meta_pages"] = nextMetaPages
	}

	return next
}

func EnrichSearchResponseWithResolvedUrls(payload interface{}, baseUrl string) interface{} {
	m, ok := payload.(map[string]interface{})
	if !ok {
		return payload
	}

	illusts, ok := m["illusts"].([]interface{})
	if !ok {
		return payload
	}

	nextIllusts := make([]interface{}, len(illusts))
	for i, item := range illusts {
		nextIllusts[i] = enrichIllust(item, baseUrl)
	}

	next := make(map[string]interface{})
	for k, v := range m {
		next[k] = v
	}
	next["illusts"] = nextIllusts
	return next
}

func EnrichArtworkResponseWithResolvedUrls(payload interface{}, baseUrl string) interface{} {
	m, ok := payload.(map[string]interface{})
	if !ok {
		return payload
	}

	illust, ok := m["illust"]
	if !ok {
		return payload
	}

	next := make(map[string]interface{})
	for k, v := range m {
		next[k] = v
	}
	next["illust"] = enrichIllust(illust, baseUrl)
	return next
}
