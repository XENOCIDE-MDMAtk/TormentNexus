// @hypercode/mcp-registry — runtime-safe class stubs
// Provides class stubs that @hypercode/core uses as both types and values.
// Real implementations live in @hypercode/core runtime.

export class Registry {
  private _servers: Map<string, any> = new Map();
  register(_name: string, _server: any): void {}
  unregister(_name: string): void {}
  get(_name: string): any { return undefined; }
  list(): any[] { return []; }
  async execute(_input: any): Promise<any> { return { result: '[Registry stub]' }; }
}
