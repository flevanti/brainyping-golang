create table checks
(
    _id                char(50)      not null,
    checkid            varchar(50)   not null,
    owneruid           varchar(50)   not null,
    name               varchar(50)   not null,
    namefriendly       varchar(1000) null,
    frequency          int           not null,
    type               varchar(50)   not null,
    subtype            varchar(50)   null,
    regions            varchar(1000) not null,
    enabled            tinyint(1)    not null,
    host               varchar(1000) not null,
    port               int           not null,
    startschedtimeunix int           not null,
    responsestring     varchar(100)  null,
    httpbody           varchar(1000) null,
    useragent          varchar(1000) null,
    httpheaders        varchar(1000) null,
    httpstatuscodeok   varchar(100)  null,
    createdunix        int           not null,
    updatedunix        int           null
);

create table settings
(
    `key`       varchar(50)   not null,
    value       varchar(1000) not null,
    description varchar(500)  null
);


