export namespace git {
	
	export class DiffLine {
	    op: string;
	    text: string;
	    oldLineNo: number;
	    newLineNo: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.op = source["op"];
	        this.text = source["text"];
	        this.oldLineNo = source["oldLineNo"];
	        this.newLineNo = source["newLineNo"];
	    }
	}
	export class Hunk {
	    oldStart: number;
	    newStart: number;
	    header: string;
	    lines: DiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new Hunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldStart = source["oldStart"];
	        this.newStart = source["newStart"];
	        this.header = source["header"];
	        this.lines = this.convertValues(source["lines"], DiffLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileDiff {
	    oldPath: string;
	    newPath: string;
	    hunks: Hunk[];
	    binary: boolean;
	    preamble: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldPath = source["oldPath"];
	        this.newPath = source["newPath"];
	        this.hunks = this.convertValues(source["hunks"], Hunk);
	        this.binary = source["binary"];
	        this.preamble = source["preamble"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileStatus {
	    Path: string;
	    IndexStatus: number;
	    WorkStatus: number;
	    Staged: boolean;
	    Unstaged: boolean;
	    Untracked: boolean;
	    Conflicted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.IndexStatus = source["IndexStatus"];
	        this.WorkStatus = source["WorkStatus"];
	        this.Staged = source["Staged"];
	        this.Unstaged = source["Unstaged"];
	        this.Untracked = source["Untracked"];
	        this.Conflicted = source["Conflicted"];
	    }
	}
	
	export class PatchLine {
	    op: string;
	    text: string;
	    selected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.op = source["op"];
	        this.text = source["text"];
	        this.selected = source["selected"];
	    }
	}
	export class PatchHunk {
	    oldStart: number;
	    newStart: number;
	    lines: PatchLine[];
	
	    static createFrom(source: any = {}) {
	        return new PatchHunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldStart = source["oldStart"];
	        this.newStart = source["newStart"];
	        this.lines = this.convertValues(source["lines"], PatchLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class DiffResult {
	    files: git.FileDiff[];
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], git.FileDiff);
	        this.raw = source["raw"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RepoInfo {
	    root: string;
	    branch: string;
	    ahead: number;
	    behind: number;
	    hasUpstream: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RepoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.branch = source["branch"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.hasUpstream = source["hasUpstream"];
	    }
	}
	export class UndoState {
	    canUndo: boolean;
	    canRedo: boolean;
	    undoDesc: string;
	    redoDesc: string;
	
	    static createFrom(source: any = {}) {
	        return new UndoState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canUndo = source["canUndo"];
	        this.canRedo = source["canRedo"];
	        this.undoDesc = source["undoDesc"];
	        this.redoDesc = source["redoDesc"];
	    }
	}

}

export namespace settings {
	
	export class Settings {
	    fontSize: number;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fontSize = source["fontSize"];
	        this.language = source["language"];
	    }
	}

}

