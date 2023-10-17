import { db } from "./db";

(async() => {
    db.schema.createTable('autoreplyreply')
        .ifNotExists()
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('reply', 'text', col => col.notNull())
        .addColumn('lastUsed', 'integer', col => col.defaultTo(0))
        .addColumn('ignoreCooldown', 'boolean', col => col.defaultTo(false))
        .execute();

    db.schema.createTable('autoreplyterm')
        .ifNotExists()
        .addColumn('id', 'integer', col => col.primaryKey().autoIncrement())
        .addColumn('term', 'text', col => col.notNull())
        .addColumn('replyId', 'integer', col => col.notNull().references('autoreplyreply.id').onDelete('cascade'))
        .execute();

    db.schema.createTable('reactionrole')
        .ifNotExists()
        .addColumn('messageId', 'text', col => col.notNull())
        .addColumn('emoteId', 'text', col => col.notNull())
        .addColumn('roleId', 'text', col => col.notNull())
        .addPrimaryKeyConstraint('primary_key', ['messageId', 'emoteId'])
        .execute();

    db.schema.alterTable('autoreplyreply')
        .addColumn('ignoreCooldown', 'boolean', col => col.defaultTo(false))
        .execute()
})();