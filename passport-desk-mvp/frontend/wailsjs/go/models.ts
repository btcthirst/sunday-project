export namespace database {
	
	export class RegistrationInput {
	    citizen_id: number;
	    registration_type: string;
	    region: string;
	    district: string;
	    settlement: string;
	    street: string;
	    house_number: string;
	    apartment_number: string;
	    registration_date: string;
	    basis_document: string;
	
	    static createFrom(source: any = {}) {
	        return new RegistrationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citizen_id = source["citizen_id"];
	        this.registration_type = source["registration_type"];
	        this.region = source["region"];
	        this.district = source["district"];
	        this.settlement = source["settlement"];
	        this.street = source["street"];
	        this.house_number = source["house_number"];
	        this.apartment_number = source["apartment_number"];
	        this.registration_date = source["registration_date"];
	        this.basis_document = source["basis_document"];
	    }
	}
	export class RegistrationOutput {
	    id: number;
	    citizen_id: number;
	    registration_type: string;
	    region: string;
	    district?: string;
	    settlement: string;
	    street: string;
	    house_number: string;
	    apartment_number?: string;
	    registration_date: string;
	    deregistration_date?: string;
	    basis_document?: string;
	    is_active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RegistrationOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.citizen_id = source["citizen_id"];
	        this.registration_type = source["registration_type"];
	        this.region = source["region"];
	        this.district = source["district"];
	        this.settlement = source["settlement"];
	        this.street = source["street"];
	        this.house_number = source["house_number"];
	        this.apartment_number = source["apartment_number"];
	        this.registration_date = source["registration_date"];
	        this.deregistration_date = source["deregistration_date"];
	        this.basis_document = source["basis_document"];
	        this.is_active = source["is_active"];
	    }
	}

}

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
	    active_address?: string;
	
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
	        this.active_address = source["active_address"];
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
	
	export class StatsOutput {
	    total_citizens: number;
	    total_registrations: number;
	    active_registrations: number;
	    new_this_month: number;
	
	    static createFrom(source: any = {}) {
	        return new StatsOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_citizens = source["total_citizens"];
	        this.total_registrations = source["total_registrations"];
	        this.active_registrations = source["active_registrations"];
	        this.new_this_month = source["new_this_month"];
	    }
	}

}

