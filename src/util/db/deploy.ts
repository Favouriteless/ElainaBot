import { db } from "./db";

(async function() {

    await db.schema.createTable('autoreplyreply')
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('reply', 'varchar', col => col.notNull())
        .ifNotExists()
        .execute();

    await db.schema.createTable('autoreplyterm')
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('term', 'varchar', col => col.notNull())
        .addColumn('replyId', 'integer', col => col
            .references('autoreplyreply.id')
            .onDelete('cascade')
        )
        .ifNotExists()
        .execute();
        
})();