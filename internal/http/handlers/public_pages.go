package handlers

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//go:embed public/*
var publicPageAssets embed.FS

func PrivacyPolicyPage(c *fiber.Ctx) error {
	return sendPublicPage(c, "privacy-policy.html", false)
}

func AccountDeletionPage(c *fiber.Ctx) error {
	return sendPublicPage(c, "account-deletion.html", true)
}

func PublicPageAsset(c *fiber.Ctx) error {
	name := filepath.Base(strings.TrimSpace(c.Params("name")))
	if name == "." || name == "" {
		return fiber.ErrNotFound
	}

	content, err := publicPageAssets.ReadFile("public/" + name)
	if err != nil {
		return fiber.ErrNotFound
	}

	switch filepath.Ext(name) {
	case ".css":
		c.Type("css", "utf-8")
	case ".js":
		c.Type("js", "utf-8")
	default:
		return fiber.ErrNotFound
	}
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.Send(content)
}

func sendPublicPage(c *fiber.Ctx, name string, noStore bool) error {
	content, err := publicPageAssets.ReadFile("public/" + name)
	if err != nil {
		return fmt.Errorf("read public page %s: %w", name, err)
	}

	c.Type("html", "utf-8")
	c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	if noStore {
		c.Set(fiber.HeaderCacheControl, "no-store")
		c.Set("Pragma", "no-cache")
	} else {
		c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	}
	return c.Send(content)
}
