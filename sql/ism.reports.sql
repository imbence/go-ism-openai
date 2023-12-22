create table ism.reports
(
    date    date,
    report  text,
    part    ism.report_part,
    content text,
    constraint reports_pk
        primary key (date, report, part)
);
