import { db } from "./util/db/db";

(async() => {
    await db.schema.alterTable('autoreplyreply')
        .addColumn('ignoreCooldown', 'boolean')
        .execute();

    db.updateTable('autoreplyreply')
        .set({ ignoreCooldown: 0 })
        .execute()
})();