create table ism.man_reports
(
    date                       date not null,
    intro                      text,
    overall_rank               text,
    respondents                text,
    commodities                text,
    manufacturing_pmi          text,
    new_orders                 text,
    new_orders_rank            text,
    production                 text,
    production_rank            text,
    employment                 text,
    employment_rank            text,
    supplier_deliveries        text,
    supplier_deliveries_rank   text,
    inventories                text,
    inventories_rank           text,
    customers_inventories      text,
    customers_inventories_rank text,
    prices                     text,
    prices_rank                text,
    backlog_of_orders          text,
    backlog_of_orders_rank     text,
    new_export_orders          text,
    new_export_orders_rank     text,
    imports                    text,
    imports_rank               text,
    buying_policy              text,
    constraint man_reports_pk
        primary key (date)
);
