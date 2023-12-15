package main

type ManReport struct {
	Date                     string `json:"date"`
	Intro                    string `json:"intro"`
	OverallRank              string `json:"overall_rank"`
	Respondents              string `json:"respondents"`
	Commodities              string `json:"commodities"`
	ManufacturingPmi         string `json:"manufacturing_pmi"`
	NewOrders                string `json:"new_orders"`
	NewOrdersRank            string `json:"new_orders_rank"`
	Production               string `json:"production"`
	ProductionRank           string `json:"production_rank"`
	Employment               string `json:"employment"`
	EmploymentRank           string `json:"employment_rank"`
	SupplierDeliveries       string `json:"supplier_deliveries"`
	SupplierDeliveriesRank   string `json:"supplier_deliveries_rank"`
	Inventories              string `json:"inventories"`
	InventoriesRank          string `json:"inventories_rank"`
	CustomersInventories     string `json:"customers_inventories"`
	CustomersInventoriesRank string `json:"customers_inventories_rank"`
	Prices                   string `json:"prices"`
	PricesRank               string `json:"prices_rank"`
	BacklogOfOrders          string `json:"backlog_of_orders"`
	BacklogOfOrdersRank      string `json:"backlog_of_orders_rank"`
	NewExportOrders          string `json:"new_export_orders"`
	NewExportOrdersRank      string `json:"new_export_orders_rank"`
	Imports                  string `json:"imports"`
	ImportsRank              string `json:"imports_rank"`
	BuyingPolicy             string `json:"buying_policy"`
}

type AiRequest []struct {
	ID      string   `json:"id"`
	Content string   `json:"content"`
	Target  []string `json:"target"`
}
