package grpc

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/maskrapp/api/internal/domains"
	"github.com/maskrapp/api/internal/global"
	stubs "github.com/maskrapp/api/internal/pb/main_api/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type mainApiServiceImpl struct {
	db      *gorm.DB
	domains *domains.Domains
	stubs.UnimplementedMainAPIServiceServer
}

func NewMainAPIService(ctx global.Context) stubs.MainAPIServiceServer {
	return &mainApiServiceImpl{
		db:      ctx.Instances().Gorm,
		domains: ctx.Instances().Domains,
	}
}

func (b *mainApiServiceImpl) CheckMask(ctx context.Context, request *stubs.CheckMaskRequest) (*stubs.CheckMaskResponse, error) {
	split := strings.Split(request.MaskAddress, "@")
	if len(split) != 2 {
		return nil, status.New(codes.InvalidArgument, "invalid mask address").Err()
	}

	if _, err := b.domains.Get(split[1]); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid mask domain").Err()
	}

	var result struct {
		Found bool
	}
	err := b.db.Raw("SELECT EXISTS(SELECT 1 FROM masks WHERE mask = ?) AS found",
		request.MaskAddress).Scan(&result).Error
	if err != nil {
		logrus.Errorf("db error: %v", err)
		return nil, status.New(codes.Unavailable, err.Error()).Err()
	}
	if !result.Found {
		return &stubs.CheckMaskResponse{Valid: false}, status.New(codes.NotFound, "mask not found").Err()
	}
	return &stubs.CheckMaskResponse{Valid: result.Found}, nil
}
func (b *mainApiServiceImpl) GetMask(ctx context.Context, request *stubs.GetMaskRequest) (*stubs.GetMaskResponse, error) {

	split := strings.Split(request.MaskAddress, "@")
	if len(split) != 2 {
		return nil, status.New(codes.InvalidArgument, "invalid mask address").Err()
	}

	if _, err := b.domains.Get(split[1]); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid mask domain").Err()
	}

	var res struct {
		Email   string
		Enabled bool
	}

	err := b.db.Table("masks").Select("masks.enabled, emails.email").Joins("inner join emails on emails.id = masks.forward_to").Where("masks.mask = ?", request.MaskAddress).Find(&res).Error
	if err != nil {
		logrus.Errorf("db error: %v", err)
		return nil, status.New(codes.Unavailable, err.Error()).Err()
	}
	return &stubs.GetMaskResponse{
		Email:   res.Email,
		Enabled: res.Enabled,
	}, nil
}
func (b *mainApiServiceImpl) IncrementForwardedCount(ctx context.Context, request *stubs.IncrementForwardedCountRequest) (*emptypb.Empty, error) {

	split := strings.Split(request.MaskAddress, "@")
	if len(split) != 2 {
		return nil, status.New(codes.InvalidArgument, "invalid mask address").Err()
	}

	if _, err := b.domains.Get(split[1]); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid mask domain").Err()
	}

	err := b.db.Table("masks").Where("mask = ?", request.MaskAddress).Updates(map[string]interface{}{"messages_received": gorm.Expr("messages_received + ?", 1), "messages_forwarded": gorm.Expr("messages_forwarded + ?", 1)}).Error

	if err != nil {
		logrus.Errorf("db error: %v", err)
		return &empty.Empty{}, status.New(codes.Unavailable, err.Error()).Err()
	}

	return &emptypb.Empty{}, nil
}
func (b *mainApiServiceImpl) IncrementReceivedCount(ctx context.Context, request *stubs.IncrementReceivedCountRequest) (*emptypb.Empty, error) {

	split := strings.Split(request.MaskAddress, "@")
	if len(split) != 2 {
		return nil, status.New(codes.InvalidArgument, "invalid mask address").Err()
	}

	if _, err := b.domains.Get(split[1]); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid mask domain").Err()
	}

	err := b.db.Table("masks").Where("mask = ?", request.MaskAddress).UpdateColumn("messages_received", gorm.Expr("messages_received + ?", 1)).Error
	if err != nil {
		logrus.Errorf("db error: %v", err)
		return &empty.Empty{}, status.New(codes.Unavailable, err.Error()).Err()
	}
	return &empty.Empty{}, nil
}
