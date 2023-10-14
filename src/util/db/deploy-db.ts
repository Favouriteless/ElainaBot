import { db } from "./db";

(async() => {
    db.schema.createTable('autoreplyreply')
        .ifNotExists()
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('reply', 'text', col => col.notNull())
        .addColumn('lastUsed', 'integer', col => col.defaultTo(0))
        .execute();

    db.schema.createTable('autoreplyterm')
        .ifNotExists()
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('term', 'text', col => col.notNull())
        .addColumn('replyId', 'integer', col => col.notNull().references('autoreplyreply.id').onDelete('cascade'))
        .execute();
})();