package usecase

import (
	"mini-pos/dto"
	"mini-pos/repository"
	"mini-pos/util"
	"time"

	"github.com/labstack/echo/v4"
)

type TransactionUseCase interface {
	Insert(echo.Context) (dto.TransactionPayload, []dto.ValidationMessage, error)
	GetAll(echo.Context) ([]dto.TransactionResponse, error)
	GetDetailByID(echo.Context, uint) (dto.TransactionDetailResponse, error)
}

type transactionUseCase struct {
	transactionRepository repository.TransactionRepository
	productRepository     repository.ProductRepository
}

func InitTransactionUseCase(transactionRepository repository.TransactionRepository, productRepository repository.ProductRepository) TransactionUseCase {
	return &transactionUseCase{
		transactionRepository: transactionRepository,
		productRepository:     productRepository,
	}
}

func (uc *transactionUseCase) Insert(ctx echo.Context) (data dto.TransactionPayload, invalidParameter []dto.ValidationMessage, err error) {

	if err = ctx.Bind(&data); err != nil {
		return
	}

	// validation
	// TODO: validate payload
	if data.OrderNumber == "" {
		invalidParameter = append(invalidParameter, dto.ValidationMessage{Parameter: "order_number", Message: "order_number is required"})
	}

	if len(invalidParameter) > 0 {
		return
	}

	// setup data
	data.UserID = util.GetSessionByName(ctx, "user_id").(uint)
	data.OrderDate = time.Now()
	data.PaymentDate = time.Now()

	var transaction dto.Transaction
	if transaction, err = uc.transactionRepository.Insert(data.ToModel()); err != nil {
		return
	}

	for _, detail := range data.TransactionDetail {
		detail.TransactionID = transaction.Id

		// get data product from outlet product
		var product dto.OutletProduct
		if product, err = uc.productRepository.GetOutletProductByID(int(detail.OutletProductID)); err != nil {
			return
		}

		detail.ProductName = product.Product.Name
		detail.Price = product.Price

		// insert transaction detail
		if _, err = uc.transactionRepository.InsertDetail(detail.ToModel()); err != nil {
			return
		}
	}

	return

}

func (uc *transactionUseCase) GetAll(ctx echo.Context) (data []dto.TransactionResponse, err error) {
	var filter dto.Transaction
	if err = ctx.Bind(&filter); err != nil {
		return
	}

	// get user id from session
	filter.UserID = util.GetSessionByName(ctx, "user_id").(uint)

	pagination := dto.InitPagination()
	if err = ctx.Bind(&pagination); err != nil {
		return
	}

	var transactions []dto.Transaction
	if transactions, err = uc.transactionRepository.GetAll(filter, pagination); err != nil {
		return
	}

	for _, transaction := range transactions {
		data = append(data, dto.TransactionResponse{
			TransactionID:  transaction.Id,
			OutletID:       transaction.OutletID,
			UserID:         transaction.UserID,
			OrderNumber:    transaction.OrderNumber,
			OrderDate:      transaction.OrderDate,
			OrderNominal:   transaction.OrderNominal,
			PaymentNumber:  transaction.PaymentNumber,
			PaymentDate:    transaction.PaymentDate,
			PaymentNominal: transaction.PaymentNominal,
			PaymentMethod:  transaction.PaymentMethod,
			PaymentNote:    transaction.PaymentNote,
			IsStatus:       transaction.IsStatus,
		})
	}

	return
}

func (uc *transactionUseCase) GetDetailByID(ctx echo.Context, id uint) (data dto.TransactionDetailResponse, err error) {
	var filter dto.TransactionPayload
	if err = ctx.Bind(&filter); err != nil {
		return
	}

	// set filter transaction id from parameter
	filter.TransactionID = id
	// get user id from session
	// filter.UserID = util.GetSessionByName(ctx, "user_id").(uint)

	// get transaction
	var transaction dto.Transaction
	if transaction, err = uc.transactionRepository.GetByID(id); err != nil {
		return
	}
	data.Transaction = dto.TransactionResponse{
		TransactionID:  transaction.Id,
		OutletID:       transaction.OutletID,
		UserID:         transaction.UserID,
		OrderNumber:    transaction.OrderNumber,
		OrderDate:      transaction.OrderDate,
		OrderNominal:   transaction.OrderNominal,
		PaymentNumber:  transaction.PaymentNumber,
		PaymentDate:    transaction.PaymentDate,
		PaymentNominal: transaction.PaymentNominal,
		PaymentMethod:  transaction.PaymentMethod,
		PaymentNote:    transaction.PaymentNote,
		IsStatus:       transaction.IsStatus,
	}

	// get detail transaction
	var detail_transactions []dto.TransactionDetail
	if detail_transactions, err = uc.transactionRepository.GetDetail(filter); err != nil {
		return
	}
	for _, detail_transaction := range detail_transactions {
		data.TransactionDetail = append(data.TransactionDetail, dto.TransactionDetailPayload{
			TransactionID:   detail_transaction.Id,
			OutletProductID: detail_transaction.OutletProductID,
			ProductName:     detail_transaction.ProductName,
			Quantity:        detail_transaction.Quantity,
			Price:           detail_transaction.Price,
		})
	}

	return
}