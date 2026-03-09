package parsing

const (
	infoAnchor     = "Сводная информация по субсчёту Клиента"
	positionAnchor = "Отчёт об остатках ценных бумаг"
	cashFlowAnchor = "Движение ценных бумаг"
	endingAnchor   = "Заключенные в отчетном периоде сделки"
)

// CashFlowOperationType - тип операции движения денежных средств
type CashFlowOperationType string

const (
	CashFlowTypeDeposit          CashFlowOperationType = "DEPOSIT"
	CashFlowTypeWithdrawal       CashFlowOperationType = "WITHDRAWAL"
	CashFlowTypeSecurityPurchase CashFlowOperationType = "SECURITY_PURCHASE"
	CashFlowTypeSecuritySale     CashFlowOperationType = "SECURITY_SALE"
	CashFlowTypeDividend         CashFlowOperationType = "DIVIDEND"
	CashFlowTypeTax              CashFlowOperationType = "TAX"
	CashFlowTypeFee              CashFlowOperationType = "FEE"
	CashFlowTypeOther            CashFlowOperationType = "OTHER"
)

// MapCashFlowOperationType преобразует текст из отчета в тип операции
func MapCashFlowOperationType(reportText string) CashFlowOperationType {
	switch reportText {
	case "Зачисление денежных средств":
		return CashFlowTypeDeposit
	case "Списание денежных средств":
		return CashFlowTypeWithdrawal
	case "Сальдо расчетов по сделкам с ценными бумагами":
		return CashFlowTypeSecurityPurchase
	case "Выплата дивидендов":
		return CashFlowTypeDividend
	case "Удержание налога":
		return CashFlowTypeTax
	case "Вознаграждение брокера", "Комиссия":
		return CashFlowTypeFee
	default:
		return CashFlowTypeOther
	}
}
