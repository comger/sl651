export namespace simulator {
	
	export class SimulationParams {
	    Type: string;
	    Base: number;
	    Min: number;
	    Max: number;
	    Step: number;
	    Scale: number;
	
	    static createFrom(source: any = {}) {
	        return new SimulationParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Base = source["Base"];
	        this.Min = source["Min"];
	        this.Max = source["Max"];
	        this.Step = source["Step"];
	        this.Scale = source["Scale"];
	    }
	}
	export class ChannelConfig {
	    Tag: number;
	    Name: string;
	    Decimals: number;
	    Algorithm: SimulationParams;
	
	    static createFrom(source: any = {}) {
	        return new ChannelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Tag = source["Tag"];
	        this.Name = source["Name"];
	        this.Decimals = source["Decimals"];
	        this.Algorithm = this.convertValues(source["Algorithm"], SimulationParams);
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
	export class DeviceConfig {
	    ID: string;
	    Name: string;
	    Password: string;
	    Interval: number;
	    HeartbeatEnabled: boolean;
	    LoginEnabled: boolean;
	    HourReportEnabled: boolean;
	    Channels: ChannelConfig[];
	
	    static createFrom(source: any = {}) {
	        return new DeviceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Password = source["Password"];
	        this.Interval = source["Interval"];
	        this.HeartbeatEnabled = source["HeartbeatEnabled"];
	        this.LoginEnabled = source["LoginEnabled"];
	        this.HourReportEnabled = source["HourReportEnabled"];
	        this.Channels = this.convertValues(source["Channels"], ChannelConfig);
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
	
	export class SimulatorConfig {
	    ServerAddr: string;
	    Devices: DeviceConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SimulatorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ServerAddr = source["ServerAddr"];
	        this.Devices = this.convertValues(source["Devices"], DeviceConfig);
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

