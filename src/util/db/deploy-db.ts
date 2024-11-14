import { db } from "./db";

(async() => {
    
    db.schema.createTable('macro')
        .ifNotExists()
        .addColumn('macro', 'text', col => col.primaryKey())
        .addColumn('reply', 'text', col => col.notNull())
        .execute();

})();
