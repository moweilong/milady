package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/moweilong/milady/pkg/copier"
	"github.com/moweilong/milady/pkg/gin/middleware"
	"github.com/moweilong/milady/pkg/gin/response"
	"github.com/moweilong/milady/pkg/logger"
	"github.com/moweilong/milady/pkg/utils"

	"github.com/moweilong/milady/internal/cache"
	"github.com/moweilong/milady/internal/dao"
	"github.com/moweilong/milady/internal/database"
	"github.com/moweilong/milady/internal/ecode"
	"github.com/moweilong/milady/internal/model"
	"github.com/moweilong/milady/internal/types"
)

var _ UserExampleHandler = (*userExampleHandler)(nil)

// UserExampleHandler defining the handler interface
type UserExampleHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type userExampleHandler struct {
	iDao dao.UserExampleDao
}

// NewUserExampleHandler creating the handler interface
func NewUserExampleHandler() UserExampleHandler {
	return &userExampleHandler{
		iDao: dao.NewUserExampleDao(
			database.GetDB(), // todo show db driver name here
			cache.NewUserExampleCache(database.GetCacheType()),
		),
	}
}

// Create a new userExample
// @Summary Create a new userExample
// @Description Creates a new userExample entity using the provided data in the request body.
// @Tags userExample
// @Accept json
// @Produce json
// @Param data body types.CreateUserExampleRequest true "userExample information"
// @Success 200 {object} types.CreateUserExampleReply{}
// @Router /api/v1/userExample [post]
// @Security BearerAuth
func (h *userExampleHandler) Create(c *gin.Context) {
	form := &types.CreateUserExampleRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	userExample := &model.UserExample{}
	err = copier.Copy(userExample, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateUserExample)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, userExample)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": userExample.ID})
}

// DeleteByID delete a userExample by id
// @Summary Delete a userExample by id
// @Description Deletes a existing userExample identified by the given id in the path.
// @Tags userExample
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteUserExampleByIDReply{}
// @Router /api/v1/userExample/{id} [delete]
// @Security BearerAuth
func (h *userExampleHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getUserExampleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.DeleteByID(ctx, id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// UpdateByID update a userExample by id
// @Summary Update a userExample by id
// @Description Updates the specified userExample by given id in the path, support partial update.
// @Tags userExample
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateUserExampleByIDRequest true "userExample information"
// @Success 200 {object} types.UpdateUserExampleByIDReply{}
// @Router /api/v1/userExample/{id} [put]
// @Security BearerAuth
func (h *userExampleHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getUserExampleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateUserExampleByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	userExample := &model.UserExample{}
	err = copier.Copy(userExample, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDUserExample)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, userExample)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a userExample by id
// @Summary Get a userExample by id
// @Description Gets detailed information of a userExample specified by the given id in the path.
// @Tags userExample
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetUserExampleByIDReply{}
// @Router /api/v1/userExample/{id} [get]
// @Security BearerAuth
func (h *userExampleHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getUserExampleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	userExample, err := h.iDao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByID not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Error(c, ecode.NotFound)
		} else {
			logger.Error("GetByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Output(c, ecode.InternalServerError.ToHTTPCode())
		}
		return
	}

	data := &types.UserExampleObjDetail{}
	err = copier.Copy(data, userExample)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDUserExample)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"userExample": data})
}

// List get a paginated list of userExamples by custom conditions
// @Summary Get a paginated list of userExamples by custom conditions
// @Description Returns a paginated list of userExample based on query filters, including page number and size.
// @Tags userExample
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListUserExamplesReply{}
// @Router /api/v1/userExample/list [post]
// @Security BearerAuth
func (h *userExampleHandler) List(c *gin.Context) {
	form := &types.ListUserExamplesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	userExamples, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertUserExamples(userExamples)
	if err != nil {
		response.Error(c, ecode.ErrListUserExample)
		return
	}

	response.Success(c, gin.H{
		"userExamples": data,
		"total":        total,
	})
}

func getUserExampleIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertUserExample(userExample *model.UserExample) (*types.UserExampleObjDetail, error) {
	data := &types.UserExampleObjDetail{}
	err := copier.Copy(data, userExample)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertUserExamples(fromValues []*model.UserExample) ([]*types.UserExampleObjDetail, error) {
	toValues := []*types.UserExampleObjDetail{}
	for _, v := range fromValues {
		data, err := convertUserExample(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
