package costco

// GraphQL Queries - Separated for maintainability

const OnlineOrdersQuery = `query getOnlineOrders($startDate:String!, $endDate:String!, $pageNumber:Int , $pageSize:Int, $warehouseNumber:String! ){
	getOnlineOrders(startDate:$startDate, endDate:$endDate, pageNumber : $pageNumber, pageSize :  $pageSize, warehouseNumber :  $warehouseNumber) {
		pageNumber
		pageSize
		totalNumberOfRecords
		bcOrders {
			orderHeaderId
			orderPlacedDate : orderedDate
			orderNumber : sourceOrderNumber 
			orderTotal
			warehouseNumber
			status
			emailAddress
			orderCancelAllowed
			orderPaymentFailed : orderPaymentEditAllowed
			orderReturnAllowed
			orderLineItems {
				orderLineItemCancelAllowed
				orderLineItemId
				orderReturnAllowed
				itemId
				itemNumber
				itemTypeId
				lineNumber
				itemDescription
				deliveryDate
				warehouseNumber
				status
				orderStatus
				parentOrderLineItemId
				isFSAEligible
				shippingType
				shippingTimeFrame
				isShipToWarehouse
				carrierItemCategory
				carrierContactPhone
				programTypeId
				isBuyAgainEligible
				scheduledDeliveryDate
				scheduledDeliveryDateEnd
				configuredItemData
				shipment {
					shipmentId             
					orderHeaderId
					orderShipToId 
					lineNumber 
					orderNumber
					shippingType 
					shippingTimeFrame 
					shippedDate 
					packageNumber 
					trackingNumber 
					trackingSiteUrl 
					carrierName         
					estimatedArrivalDate 
					deliveredDate 
					isDeliveryDelayed 
					isEstimatedArrivalDateEligible 
					statusTypeId 
					status 
					pickUpReadyDate
					pickUpCompletedDate
					reasonCode
					trackingEvent {
						event
						carrierName
						eventDate
						estimatedDeliveryDate
						scheduledDeliveryDate
						trackingNumber
					}
				}
			}
		}
	}
}`

const ReceiptsQuery = `query receiptsWithCounts($startDate: String!, $endDate: String!,$documentType:String!,$documentSubType:String!) {
	receiptsWithCounts(startDate: $startDate, endDate: $endDate,documentType:$documentType,documentSubType:$documentSubType) {
		inWarehouse
		gasStation
		carWash
		gasAndCarWash
		receipts{
			warehouseName 
			receiptType  
			documentType 
			transactionDateTime 
			transactionBarcode 
			warehouseName 
			transactionType 
			total 
			totalItemCount
			itemArray {  
				itemNumber
			}
			tenderArray {   
				tenderTypeCode
				tenderDescription
				amountTender
			}
			couponArray {  
				upcnumberCoupon
			}  
		}
	}
}`

const ReceiptDetailQuery = `query receiptsWithCounts($barcode: String!,$documentType:String!) {
	receiptsWithCounts(barcode: $barcode,documentType:$documentType) {
		receipts{
			warehouseName
			receiptType 
			documentType 
			transactionDateTime 
			transactionDate 
			companyNumber  
			warehouseNumber 
			operatorNumber  
			warehouseName  
			warehouseShortName   
			registerNumber  
			transactionNumber  
			transactionType
			transactionBarcode  
			total 
			warehouseAddress1 
			warehouseAddress2 
			warehouseCity 
			warehouseState 
			warehouseCountry 
			warehousePostalCode
			totalItemCount 
			subTotal 
			taxes
			total 
			invoiceNumber
			sequenceNumber
			itemArray {  
				itemNumber 
				itemDescription01 
				frenchItemDescription1 
				itemDescription02 
				frenchItemDescription2 
				itemIdentifier 
				itemDepartmentNumber
				unit 
				amount 
				taxFlag 
				merchantID 
				entryMethod
				transDepartmentNumber
				fuelUnitQuantity
				fuelGradeCode
				fuelUnitQuantity
				itemUnitPriceAmount
				fuelUomCode
				fuelUomDescription
				fuelUomDescriptionFr
				fuelGradeDescription
				fuelGradeDescriptionFr
			}  
			tenderArray {   
				tenderTypeCode
				tenderSubTypeCode
				tenderDescription    
				amountTender    
				displayAccountNumber   
				sequenceNumber   
				approvalNumber   
				responseCode 
				tenderTypeName 
				transactionID   
				merchantID   
				entryMethod
				tenderAcctTxnNumber  
				tenderAuthorizationCode  
				tenderTypeName
				tenderTypeNameFr
				tenderEntryMethodDescription
				walletType
				walletId
				storedValueBucket
			}    
			subTaxes {      
				tax1      
				tax2      
				tax3     
				tax4     
				aTaxPercent     
				aTaxLegend     
				aTaxAmount
				aTaxPrintCode
				aTaxPrintCodeFR     
				aTaxIdentifierCode     
				bTaxPercent    
				bTaxLegend     
				bTaxAmount
				bTaxPrintCode
				bTaxPrintCodeFR     
				bTaxIdentifierCode      
				cTaxPercent     
				cTaxLegend    
				cTaxAmount
				cTaxIdentifierCode           
				dTaxPercent     
				dTaxLegend     
				dTaxAmount
				dTaxPrintCode
				dTaxPrintCodeFR     
				dTaxIdentifierCode
				uTaxLegend
				uTaxAmount
				uTaxableAmount
			}   
			instantSavings   
			membershipNumber 
		}
	}
}`

// Future queries can be added here:
// const ProductSearchQuery = `...`
// const MembershipInfoQuery = `...`
// const WarehouseLocationsQuery = `...`