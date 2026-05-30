package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
	"url-shortener/internal/http/middleware"
	"url-shortener/internal/service"
)

type LinkService interface {
	CreateShortLink(ctx context.Context, rawURL string, userID *int64) (service.ShortLink, error)
	CreateCustomShortLink(ctx context.Context, rawURL string, customCode string, userID *int64) (service.ShortLink, error)
	RedirectByCode(ctx context.Context, code string, meta service.VisitMeta) (domain.Link, error)
	GetLinkStats(ctx context.Context, userID int64, code string, limit int) (service.LinkStats, error)
	ListUserLinks(ctx context.Context, userID int64) ([]service.ShortLink, error)
	UpdateLink(ctx context.Context, userID int64, currentCode string, patch service.LinkUpdate) (service.ShortLink, error)
	DeleteLink(ctx context.Context, userID int64, code string) error
	GenerateLinkQR(ctx context.Context, userID int64, code string, size int) ([]byte, error)
}

type LinkHandler struct {
	service LinkService
	logger  *zap.Logger
}

func NewLinkHandler(service LinkService, logger *zap.Logger) *LinkHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LinkHandler{
		service: service,
		logger:  logger,
	}
}

type shortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code"`
}

type shortenResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type listLinksResponse struct {
	Links []linkResponse `json:"links"`
}

type linkResponse struct {
	Code        string    `json:"code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	ClicksCount int64     `json:"clicks_count"`
	QRURL       string    `json:"qr_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type linkStatsResponse struct {
	Code        string               `json:"code"`
	ShortURL    string               `json:"short_url"`
	OriginalURL string               `json:"original_url"`
	ClicksCount int64                `json:"clicks_count"`
	Visits      []visitStatsResponse `json:"visits"`
}

type visitStatsResponse struct {
	VisitedAt time.Time `json:"visited_at"`
	IP        *string   `json:"ip"`
	UserAgent *string   `json:"user_agent"`
}

func (h *LinkHandler) Shorten(c *gin.Context) {
	h.logger.Debug("shorten handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)

	var request shortenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	if request.URL == "" {
		writeError(c, http.StatusBadRequest, "url_required", "Field url is required")
		return
	}

	var userID *int64
	if currentUserID, ok := middleware.CurrentUserID(c); ok {
		userID = &currentUserID
		h.logger.Debug("shorten request parsed", zap.Int64("user_id", currentUserID))
	} else {
		h.logger.Debug("shorten request parsed as anonymous")
	}

	var link service.ShortLink
	var err error
	if strings.TrimSpace(request.CustomCode) != "" {
		link, err = h.service.CreateCustomShortLink(c.Request.Context(), request.URL, request.CustomCode, userID)
	} else {
		link, err = h.service.CreateShortLink(c.Request.Context(), request.URL, userID)
	}
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
			return
		}

		if errors.Is(err, domain.ErrCodeAlreadyExists) {
			writeError(c, http.StatusConflict, "code_already_exists", "Short code already exists")
			return
		}

		if errors.Is(err, service.ErrInvalidURL) {
			writeError(c, http.StatusBadRequest, "invalid_url", "URL must be absolute and use http or https scheme")
			return
		}

		h.logger.Error("create short link failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	h.logger.Debug("shorten response ready",
		zap.String("code", link.Code),
		zap.String("short_url", link.ShortURL),
	)

	c.JSON(http.StatusCreated, shortenResponse{
		Code:        link.Code,
		ShortURL:    link.ShortURL,
		OriginalURL: link.OriginalURL,
	})
}

type updateLinkRequest struct {
	OriginalURL *string `json:"original_url"`
	Code        *string `json:"code"`
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token is required")
		return
	}

	currentCode := c.Param("code")

	var request updateLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	updated, err := h.service.UpdateLink(c.Request.Context(), userID, currentCode, service.LinkUpdate{
		OriginalURL: request.OriginalURL,
		Code:        request.Code,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNothingToUpdate):
			writeError(c, http.StatusBadRequest, "nothing_to_update", "At least one field must be provided")
		case errors.Is(err, service.ErrInvalidURL):
			writeError(c, http.StatusBadRequest, "invalid_url", "URL must be absolute and use http or https scheme")
		case errors.Is(err, service.ErrInvalidCode):
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
		case errors.Is(err, domain.ErrCodeAlreadyExists):
			writeError(c, http.StatusConflict, "code_already_exists", "Short code already exists")
		case errors.Is(err, domain.ErrLinkNotFound):
			writeError(c, http.StatusNotFound, "link_not_found", "Short link not found")
		default:
			h.logger.Error("update link failed",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int64("user_id", userID),
				zap.String("code", currentCode),
			)
			writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, linkResponse{
		Code:        updated.Code,
		ShortURL:    updated.ShortURL,
		OriginalURL: updated.OriginalURL,
		ClicksCount: updated.ClicksCount,
		CreatedAt:   updated.CreatedAt,
	})
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token is required")
		return
	}

	code := c.Param("code")
	if err := h.service.DeleteLink(c.Request.Context(), userID, code); err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
			return
		}
		if errors.Is(err, domain.ErrLinkNotFound) {
			writeError(c, http.StatusNotFound, "link_not_found", "Short link not found")
			return
		}

		h.logger.Error("delete link failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int64("user_id", userID),
			zap.String("code", code),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) GetLinkQR(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		h.logger.Warn("qr generation rejected: missing current user",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		writeError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token is required")
		return
	}

	code := c.Param("code")
	size, err := parseQRSize(c.Query("size"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_size", "Query size must be between 128 and 1024")
		return
	}

	h.logger.Debug("qr generation handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int64("user_id", userID),
		zap.String("code", code),
		zap.Int("size", size),
	)

	png, err := h.service.GenerateLinkQR(c.Request.Context(), userID, code, size)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCode):
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
		case errors.Is(err, service.ErrInvalidSize):
			writeError(c, http.StatusBadRequest, "invalid_size", "Query size must be between 128 and 1024")
		case errors.Is(err, domain.ErrLinkNotFound):
			writeError(c, http.StatusNotFound, "link_not_found", "Short link not found")
		case errors.Is(err, domain.ErrLinkForbidden):
			writeError(c, http.StatusForbidden, "link_forbidden", "You do not have access to this link")
		default:
			h.logger.Error("qr generation failed",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int64("user_id", userID),
				zap.String("code", code),
			)
			writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}

	c.Header("Content-Disposition", "inline; filename=\""+code+".png\"")
	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, "image/png", png)
}

func parseQRSize(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}

	size, err := strconv.Atoi(raw)
	if err != nil {
		return 0, service.ErrInvalidSize
	}

	return size, nil
}

func (h *LinkHandler) ListUserLinks(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		h.logger.Warn("list user links rejected: missing current user",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		writeError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token is required")
		return
	}

	h.logger.Debug("list user links handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int64("user_id", userID),
	)

	links, err := h.service.ListUserLinks(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list user links failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int64("user_id", userID),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	responseLinks := make([]linkResponse, 0, len(links))
	for _, link := range links {
		responseLinks = append(responseLinks, linkResponse{
			Code:        link.Code,
			ShortURL:    link.ShortURL,
			OriginalURL: link.OriginalURL,
			ClicksCount: link.ClicksCount,
			QRURL:       link.QRURL,
			CreatedAt:   link.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, listLinksResponse{Links: responseLinks})
}

func (h *LinkHandler) GetLinkStats(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		h.logger.Warn("link stats rejected: missing current user",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		writeError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token is required")
		return
	}

	code := c.Param("code")
	limit, err := parseVisitLimit(c.Query("limit"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_limit", "Query limit must be between 1 and 100")
		return
	}

	h.logger.Debug("link stats handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int64("user_id", userID),
		zap.String("code", code),
		zap.Int("limit", limit),
	)

	stats, err := h.service.GetLinkStats(c.Request.Context(), userID, code, limit)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
			return
		}

		if errors.Is(err, domain.ErrLinkNotFound) {
			writeError(c, http.StatusNotFound, "link_not_found", "Short link not found")
			return
		}

		if errors.Is(err, domain.ErrLinkForbidden) {
			writeError(c, http.StatusForbidden, "link_forbidden", "You do not have access to this link statistics")
			return
		}

		if errors.Is(err, service.ErrInvalidLimit) {
			writeError(c, http.StatusBadRequest, "invalid_limit", "Query limit must be between 1 and 100")
			return
		}

		h.logger.Error("link stats lookup failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int64("user_id", userID),
			zap.String("code", code),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	visits := make([]visitStatsResponse, 0, len(stats.Visits))
	for _, visit := range stats.Visits {
		visits = append(visits, visitStatsResponse{
			VisitedAt: visit.VisitedAt,
			IP:        visit.IP,
			UserAgent: visit.UserAgent,
		})
	}

	c.JSON(http.StatusOK, linkStatsResponse{
		Code:        stats.Code,
		ShortURL:    stats.ShortURL,
		OriginalURL: stats.OriginalURL,
		ClicksCount: stats.ClicksCount,
		Visits:      visits,
	})
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		code = strings.TrimPrefix(c.Request.URL.Path, "/")
	}

	h.logger.Debug("redirect handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("code", code),
	)

	meta := service.VisitMeta{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	link, err := h.service.RedirectByCode(c.Request.Context(), code, meta)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			h.logger.Warn("redirect rejected: invalid code", zap.String("code", code))
			writeError(c, http.StatusBadRequest, "invalid_code", "Code contains unsupported characters")
			return
		}

		if errors.Is(err, domain.ErrLinkNotFound) {
			h.logger.Warn("redirect rejected: link not found", zap.String("code", code))
			writeError(c, http.StatusNotFound, "link_not_found", "Short link not found")
			return
		}

		h.logger.Error("redirect link lookup failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("code", code),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	h.logger.Debug("redirect response ready",
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
		zap.String("url_scheme", linkURLScheme(link.OriginalURL)),
		zap.String("url_host", linkURLHost(link.OriginalURL)),
	)

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func writeError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func linkURLScheme(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsed.Scheme
}

func linkURLHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsed.Host
}

func parseVisitLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > service.MaxVisitLimit {
		return 0, service.ErrInvalidLimit
	}

	return limit, nil
}
