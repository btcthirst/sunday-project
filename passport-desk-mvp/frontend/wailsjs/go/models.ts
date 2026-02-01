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

export namespace services {
	
	export class CitizenInput {
	    last_name: string;
	    first_name: string;
	    middle_name: string;
	    birth_date: string;
	    passport_series: string;
	    passport_number: string;
	    tax_number: string;
	    gender: string;
	    birth_place: string;
	    phone: string;
	    email: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CitizenInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.last_name = source["last_name"];
	        this.first_name = source["first_name"];
	        this.middle_name = source["middle_name"];
	        this.birth_date = source["birth_date"];
	        this.passport_series = source["passport_series"];
	        this.passport_number = source["passport_number"];
	        this.tax_number = source["tax_number"];
	        this.gender = source["gender"];
	        this.birth_place = source["birth_place"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.notes = source["notes"];
	    }
	}
	export class CitizenOutput {
	    id: number;
	    last_name: string;
	    first_name: string;
	    middle_name: string;
	    full_name: string;
	    birth_date: string;
	    passport_series: string;
	    passport_number: string;
	    passport_masked: string;
	    tax_number: string;
	    tax_number_masked: string;
	    gender: string;
	    gender_display: string;
	    birth_place: string;
	    phone: string;
	    email: string;
	    notes: string;
	    deleted: boolean;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CitizenOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.last_name = source["last_name"];
	        this.first_name = source["first_name"];
	        this.middle_name = source["middle_name"];
	        this.full_name = source["full_name"];
	        this.birth_date = source["birth_date"];
	        this.passport_series = source["passport_series"];
	        this.passport_number = source["passport_number"];
	        this.passport_masked = source["passport_masked"];
	        this.tax_number = source["tax_number"];
	        this.tax_number_masked = source["tax_number_masked"];
	        this.gender = source["gender"];
	        this.gender_display = source["gender_display"];
	        this.birth_place = source["birth_place"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.notes = source["notes"];
	        this.deleted = source["deleted"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CitizenListResult {
	    items: CitizenOutput[];
	    total: number;
	    page: number;
	    limit: number;
	    total_pages: number;
	
	    static createFrom(source: any = {}) {
	        return new CitizenListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], CitizenOutput);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.limit = source["limit"];
	        this.total_pages = source["total_pages"];
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

