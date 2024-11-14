import { Insertable, Selectable, Updateable } from "kysely";

// ----------------------------- DATABASE -----------------------------

export interface Database {
	macro: MacroTable;
}

// ------------------------------ TABLES ------------------------------

export interface MacroTable {
	macro: string;
	reply: string;
}
export type Macro = Selectable<MacroTable>;
export type NewMacro = Insertable<MacroTable>;
export type UpdateMacro = Updateable<MacroTable>;