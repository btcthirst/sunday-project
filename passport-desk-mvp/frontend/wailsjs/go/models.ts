export namespace main {
	
	export class OperatorInfo {
	    id: number;
	    username: string;
	    full_name: string;
	
	    static createFrom(source: any = {}) {
	        return new OperatorInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.full_name = source["full_name"];
	    }
	}

}

