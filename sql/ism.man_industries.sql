create table ism.man_industries
(
    id        integer not null,
    shortname text    not null,
    name      text,
    constraint man_industries_pk
        primary key (id, shortname)
            deferrable
);
