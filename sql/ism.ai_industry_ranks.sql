create table ism.ai_industry_ranks
(
    date     date not null,
    industry text not null,
    rank     integer,
    comment  text,
    constraint ai_industry_ranks_pk
        primary key (date, industry)
);
