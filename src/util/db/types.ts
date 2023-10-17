import { Generated, Insertable, Selectable, Updateable } from "kysely";

// ----------------------------- DATABASE -----------------------------

export interface Database {
	autoreplyreply: AutoreplyReplyTable;
	autoreplyterm: AutoreplyTermTable;
	reactionrole: ReactionRoleTable;
}

// ------------------------------ TABLES ------------------------------

export interface AutoreplyReplyTable {
	id: Generated<number>;
	reply: string;
	lastUsed: Generated<number>;
	ignoreCooldown: number;
}
export type AutoreplyReply = Selectable<AutoreplyReplyTable>;
export type NewAutoreplyReply = Insertable<AutoreplyReplyTable>;
export type AutoreplyReplyUpdate = Updateable<AutoreplyReplyTable>;


export interface AutoreplyTermTable {
	id: Generated<number>;
	term: string;
	replyId: number;
}
export type AutoreplyTerm = Selectable<AutoreplyTermTable>;
export type NewAutoreplyTerm = Insertable<AutoreplyTermTable>;
export type AutoreplyTermUpdate = Updateable<AutoreplyTermTable>;

export interface ReactionRoleTable {
	messageId: string;
	emoteId: string;
	roleId: string;
}

export type ReactionRole = Selectable<ReactionRoleTable>;
export type NewReactionRole = Insertable<ReactionRoleTable>;
export type ReactionRoleUpdate = Updateable<ReactionRoleTable>;
