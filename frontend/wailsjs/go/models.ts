export namespace config {
	
	export class Settings {
	    poll_interval_ms: number;
	    enabled: boolean;
	    direction: string;
	    start_with_windows: boolean;
	    start_minimized: boolean;
	    write_back_to_clipboard: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.poll_interval_ms = source["poll_interval_ms"];
	        this.enabled = source["enabled"];
	        this.direction = source["direction"];
	        this.start_with_windows = source["start_with_windows"];
	        this.start_minimized = source["start_minimized"];
	        this.write_back_to_clipboard = source["write_back_to_clipboard"];
	    }
	}

}

export namespace main {
	
	export class Status {
	    enabled: boolean;
	    running: boolean;
	    counter: number;
	    clipboard_head: string;
	    last_event?: service.Event;
	    // Go type: time
	    last_updated: any;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.running = source["running"];
	        this.counter = source["counter"];
	        this.clipboard_head = source["clipboard_head"];
	        this.last_event = this.convertValues(source["last_event"], service.Event);
	        this.last_updated = this.convertValues(source["last_updated"], null);
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

export namespace service {
	
	export class Event {
	    kind: string;
	    // Go type: time
	    when: any;
	    source?: string;
	    target?: string;
	    direction?: string;
	    skip_reason?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.when = this.convertValues(source["when"], null);
	        this.source = source["source"];
	        this.target = source["target"];
	        this.direction = source["direction"];
	        this.skip_reason = source["skip_reason"];
	        this.error = source["error"];
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

