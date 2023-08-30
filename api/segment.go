package api

import (
	"assignment/domain"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

type Controller struct {
	SegmentService domain.SegmentService
	Log            *slog.Logger
}

func New(service domain.SegmentService) *Controller {
	return &Controller{
		SegmentService: service,
	}
}

func (c *Controller) CreateSegment(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed reading body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body struct {
		Segment string `json:"segment"`
	}

	if err := json.Unmarshal(rawBody, &body); err != nil {
		c.Log.ErrorContext(ctx, "failed unmarshaling body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.SegmentService.CreateSegment(req.Context(), body.Segment)
	if err != nil {
		var resp []byte
		if errors.Is(err, domain.ErrSegmentIsAlreadyExists) {
			resp, _ = json.Marshal(map[string]string{"error": "segment with this name is already exists"})
			_, err = w.Write(resp)
			if err != nil {
				c.Log.ErrorContext(ctx, "failed to write response", slog.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}
		c.Log.ErrorContext(ctx, "failed to create segment", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *Controller) DeleteSegment(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed reading body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body struct {
		Segment string `json:"segment"`
	}

	if err := json.Unmarshal(rawBody, &body); err != nil {
		c.Log.ErrorContext(ctx, "failed unmarshaling body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.SegmentService.DeleteSegment(req.Context(), body.Segment)
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			var resp []byte
			resp, _ = json.Marshal(map[string]string{"error": "can't find the segment"})
			_, err = w.Write(resp)
			if err != nil {
				c.Log.ErrorContext(ctx, "failed to write response", slog.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}
		c.Log.ErrorContext(ctx, "failed to delete segment", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) ChangeUserSegments(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed reading body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body struct {
		UserId           int      `json:"user_id"`
		SegmentsToAdd    []string `json:"segments_to_add"`
		SegmentsToDelete []string `json:"segments_to_delete"`
	}

	if err := json.Unmarshal(rawBody, &body); err != nil {
		c.Log.ErrorContext(ctx, "failed unmarshaling body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.SegmentService.ChangeUserSegments(req.Context(), body.UserId, body.SegmentsToAdd, body.SegmentsToDelete)
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) || errors.Is(err, domain.ErrUserHaveNotThisSegment) || errors.Is(err, domain.ErrUserIsAlreadyHasThisSegment) {
			errs := []string{}
			if errors.Is(err, domain.ErrSegmentNotFound) {
				errs = append(errs, "can't find the segment")
			}
			if errors.Is(err, domain.ErrUserHaveNotThisSegment) {
				errs = append(errs, "user doesn't have this segment")
			}
			if errors.Is(err, domain.ErrUserIsAlreadyHasThisSegment) {
				errs = append(errs, "user is already has this segment")
			}
			var resp []byte
			j := map[string][]string{}
			j["errors"] = errs
			resp, _ = json.Marshal(j)
			_, err = w.Write(resp)
			if err != nil {
				c.Log.ErrorContext(ctx, "failed to write response", slog.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *Controller) GetUserSegments(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed reading body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body struct {
		UserId int `json:"user_id"`
	}

	if err := json.Unmarshal(rawBody, &body); err != nil {
		c.Log.ErrorContext(ctx, "failed unmarshaling body", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	segmnets, err := c.SegmentService.GetUserSegments(req.Context(), body.UserId)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed to get user segments", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resp []byte

	type userSegments struct {
		UserId       int      `json:"user_id"`
		UserSegments []string `json:"user_segments"`
	}
	thisUserSegments := userSegments{
		UserId:       body.UserId,
		UserSegments: segmnets,
	}

	resp, err = json.Marshal(thisUserSegments)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed to marshal segments", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		c.Log.ErrorContext(ctx, "failed to write response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
