import { useState, useRef, useEffect } from 'react';

function App() {
  // --- CORE STATE ---
  const [equation, setEquation] = useState("");
  const [solutions, setSolutions] = useState([]);
  const [isSolving, setIsSolving] = useState(false);
  
  // Navigation & UI States
  const [activeSidebarTab, setActiveSidebarTab] = useState('solver');
  const [headerTab, setHeaderTab] = useState('solver'); 
  const [searchSpace, setSearchSpace] = useState("0");
  const [isLightMode, setIsLightMode] = useState(false);
  
  // Rule Engine States
  const [rules, setRules] = useState([]);
  const [showRuleForm, setShowRuleForm] = useState(false);
  const [newRuleVar, setNewRuleVar] = useState('');
  const [newRuleOp, setNewRuleOp] = useState('>');
  const [newRuleVal, setNewRuleVal] = useState('');

  const [logs, setLogs] = useState([]);
  const solutionsEndRef = useRef(null);
  const ws = useRef(null);

  // Auto-scroll for the solutions table
  useEffect(() => {
    if (headerTab === 'solver') {
      solutionsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [solutions, headerTab]);

  const handleAddRule = () => {
    if (!newRuleVar || !newRuleVal) return;
    const newRule = { id: Date.now(), variable: newRuleVar.toLowerCase().trim(), operator: newRuleOp, value: newRuleVal.trim() };
    setRules([...rules, newRule]);
    setNewRuleVar(''); setNewRuleVal(''); setShowRuleForm(false);
  };

  const handleDeleteRule = (id) => setRules(rules.filter(r => r.id !== id));
  const getConstraintsString = () => rules.map(r => `${r.variable}${r.operator}${r.value}`).join(', ');

  const startSolving = () => {
    // FAILSAFE 1: Missing Budget
    if (!equation.includes("=")) {
      setSolutions(["ERROR: Oops! You forgot the '=' sign. Please specify a budget (e.g. '= 5000')."]);
      setLogs(["> ERROR: Syntax validation failed. Missing '=' parameter."]);
      return;
    }
    
   // FAILSAFE 2: High-Degree Polynomial Shield (Allowing Quadratics now!)
  
    if (/[a-zA-Z]\s*[*\/]\s*[a-zA-Z]/.test(equation)) {
      setSolutions(["ERROR: Multi-variable multiplication (x*y) is blocked in V3."]);
      setLogs(["> ERROR: Complexity exceeds current solver threshold."]);
      return;
    }
    
    setSolutions([]);
    setLogs([]);
    setSearchSpace("0");
    setIsSolving(true);

    // Simulated Tokenizing & Extracting for Terminal Visuals
    setTimeout(() => {
      const tokens = equation.match(/([0-9.]+|[a-zA-Z]+|[+\-*/=()])/g) || [];
      setLogs([`Step 1: Tokenizing... [${tokens.map(t => `'${t}'`).join(', ')}]`]);
    }, 100);

    setTimeout(() => {
      const vars = [...new Set(equation.match(/[a-zA-Z]+/g) || [])];
      setLogs(prev => [...prev, `Step 2: Extracting Variables... Assets detected: ${vars.join(', ')}`]);
    }, 800);

    setTimeout(() => {
      const constraintStr = getConstraintsString();
      setLogs(prev => [...prev, `Step 3: Applying Rules: ${constraintStr ? constraintStr : "None"}...`]);
    }, 1500);

    // WebSocket Execution
    setTimeout(() => {
      ws.current = new WebSocket('ws://localhost:8080/ws');
      ws.current.onopen = () => { ws.current.send(JSON.stringify({ equation: equation, constraints: getConstraintsString() })); };
      ws.current.onmessage = (event) => {
        if (event.data === "FINISHED") {
          setIsSolving(false);
          setLogs(prev => [...prev, `> EXECUTION COMPLETE. Link closed.`]);
          ws.current.close();
        } else if (event.data.startsWith("STATS:")) {
          const count = event.data.split(":")[1];
          setSearchSpace(parseInt(count).toLocaleString());
        } else if (event.data.startsWith("Solution Found:")) {
          const cleanSol = event.data.replace("Solution Found: ", "");
          setSolutions((prev) => [...prev, cleanSol]);
        } else if (event.data.startsWith(">")) { 
          // NEW FIX: Route backend progress updates directly to the terminal!
          setLogs((prev) => [...prev, event.data]);
        } else {
          setSolutions((prev) => [...prev, `ERROR: ${event.data}`]);
          setLogs(prev => [...prev, `> ERROR DETECTED.`]);
        }
      };
      ws.current.onerror = () => {
        setSolutions([`ERROR: Connection failed. Ensure Go Engine is running on port 8080.`]);
        setLogs(prev => [...prev, `> FATAL ERROR: Engine Offline.`]);
        setIsSolving(false);
      };
    }, 2200); 
  };

  const getDynamicHeaders = () => {
    if (solutions.length === 0) return [];
    const validSol = solutions.find(s => !s.startsWith("ERROR"));
    if (!validSol) return [];
    return validSol.split(", ").map(pair => pair.split("=")[0]);
  };
  const tableHeaders = getDynamicHeaders();

  return (
    <div className={`min-h-screen bg-quant-bg text-quant-white font-sans flex overflow-hidden transition-all duration-500 ${isLightMode ? 'invert hue-rotate-180' : ''}`}>
      
      {/* SIDEBAR */}
      <aside className="w-64 bg-quant-bg border-r border-quant-border flex flex-col z-10">
        <div className="h-16 flex items-center px-6">
          <span className="text-xl font-bold tracking-tight text-quant-primary">Quant<span className="text-white">Solve</span></span>
        </div>
        <div className="px-6 py-4 flex items-center gap-3">
          <div className={`w-10 h-10 rounded bg-quant-panel border border-quant-border flex items-center justify-center text-lg shadow-inner ${isLightMode ? 'invert hue-rotate-180' : ''}`}>👨‍💼</div>
          <div><div className="text-sm font-bold text-white">QuantSolve</div><div className="text-xs text-quant-text">Quant Analyst</div></div>
        </div>
        <nav className="flex-1 px-3 py-4 space-y-1">
          <button onClick={() => {setHeaderTab('solver'); setActiveSidebarTab('solver')}} className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-semibold transition-colors ${activeSidebarTab === 'solver' ? 'bg-quant-panel text-white' : 'text-quant-text hover:text-white'}`}>
            <span className={activeSidebarTab === 'solver' ? 'text-quant-primary' : ''}>⊞</span> Workspaces
          </button>
        </nav>
      </aside>

      {/* MAIN CONTENT */}
      <main className="flex-1 flex flex-col h-screen overflow-hidden bg-[#0A0E17]">
        
        {/* HEADER */}
        <header className="h-16 border-b border-quant-border flex items-center justify-between px-8 bg-quant-bg/50 backdrop-blur-sm">
          <div className="flex gap-8 text-sm font-semibold text-quant-text">
            <button onClick={() => setHeaderTab('solver')} className={`h-16 transition-all ${headerTab === 'solver' ? 'text-white border-b-2 border-quant-primary' : 'hover:text-white'}`}>Solver Dashboard</button>
            <button onClick={() => setHeaderTab('instructions')} className={`h-16 transition-all ${headerTab === 'instructions' ? 'text-white border-b-2 border-quant-primary' : 'hover:text-white'}`}>Instructions</button>
          </div>
          
          <div className="flex items-center gap-4">
            <button onClick={() => setIsLightMode(!isLightMode)} className="flex items-center justify-center w-8 h-8 rounded bg-quant-bg border border-quant-border text-quant-text hover:text-white transition-all shadow-sm">
              <span className={isLightMode ? 'invert hue-rotate-180' : ''}>{isLightMode ? '🌙' : '☀️'}</span>
            </button>
            <button className="bg-quant-primary text-white text-xs font-bold px-4 py-2 rounded-md shadow-sm">Live Status</button>
          </div>
        </header>

        <div className="flex-1 p-8 overflow-y-auto custom-scrollbar">
          
          {/* VIEW 1: INSTRUCTIONS */}
          {headerTab === 'instructions' && (
            <div className="max-w-[1000px] mx-auto animate-in fade-in zoom-in-95 duration-300">
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-10 shadow-2xl">
                <h2 className="text-2xl font-bold text-white mb-8 border-b border-quant-border/50 pb-4 flex items-center gap-3">
                   📖 Engine Syntax & Limitations
                </h2>
                
                <div className="grid grid-cols-2 gap-12">
                  <div className="space-y-6">
                    <h3 className="text-emerald-400 font-bold uppercase tracking-widest text-xs flex items-center gap-2">✅ Valid Syntax Rules</h3>
                    <div className="space-y-4 text-gray-300 font-mono text-sm">
                      <p>● <strong>Target Budget:</strong> Must include <code className="text-emerald-300">= [number]</code> (e.g., <code className="text-emerald-300">= 5000</code>)</p>
                      <p>● <strong>Coefficients:</strong> Asset needs numeric cost. Use <code className="text-emerald-300">1x</code>, not <code className="text-emerald-300">x</code></p>
                      <p>● <strong>Decimals:</strong> Floating-points supported (e.g., <code className="text-emerald-300">10.25a</code>)</p>
                      <p>● <strong>Brackets:</strong> Multipliers supported (e.g., <code className="text-emerald-300">2*(5x+10y)</code>)</p>
                    </div>
                  </div>

                  <div className="space-y-6">
                    <h3 className="text-rose-400 font-bold uppercase tracking-widest text-xs flex items-center gap-2">🚫 System Limitations</h3>
                    <div className="space-y-4 text-gray-300 font-mono text-sm">
                      <p>● <strong>Non-Linear:</strong> No <code className="text-rose-300">x * y</code> variable multiplication</p>
                      <p>● <strong>No Quadratics:</strong> Exponents <code className="text-rose-300">x^2</code> strictly forbidden</p>
                      <p>● <strong>Zero-Cost:</strong> Assets cannot cost <code className="text-rose-300">0x</code> (Infinity Trap)</p>
                      <p>● <strong>Negative Budgets:</strong> Budget must be strictly <code className="text-rose-300">&gt; 0</code></p>
                    </div>
                  </div>
                </div>
                
                <div className="mt-12 p-6 bg-blue-900/10 border border-blue-500/20 rounded-lg">
                  <h4 className="text-blue-400 font-bold text-xs uppercase mb-2 flex items-center gap-2">💡 Pro-Tip for Analysts</h4>
                  <p className="text-xs text-gray-400 leading-relaxed font-mono italic">
                    Large search spaces can cause latency. Use the <strong>Rule Engine</strong> on the solver dashboard to set Min/Max constraints. This prunes the mathematical tree and optimizes the Go execution path.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* VIEW 2: SOLVER DASHBOARD */}
          {headerTab === 'solver' && (
            <div className="max-w-[1400px] mx-auto space-y-6 animate-in fade-in duration-300">
              
              {/* KPIs (Optimized 2-Column Grid) */}
              <div className="grid grid-cols-2 gap-6">
                <div className="bg-[#0D1321] border border-quant-border rounded-xl p-6 shadow-lg hover:border-blue-500/50 transition-all group">
                  <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2 group-hover:text-blue-400 transition-colors">Possibilities Checked</div>
                  <div className="text-4xl font-mono text-white font-medium">
                    {isSolving ? <span className="animate-pulse">Crunching...</span> : searchSpace}
                  </div>
                </div>
                <div className="bg-[#0D1321] border border-quant-border rounded-xl p-6 shadow-lg hover:border-emerald-500/50 transition-all group">
                  <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2 group-hover:text-emerald-400 transition-colors">Solutions Found</div>
                  <div className="text-4xl font-mono text-white font-medium">
                    {solutions.length > 0 && !solutions[0].startsWith("ERROR") ? solutions.length : "0"}
                  </div>
                </div>
              </div>

              {/* EQUATION INPUT & TERMINAL */}
              <div className="grid grid-cols-3 gap-6">
                <div className="col-span-2 bg-[#0D1321] border border-quant-border rounded-xl overflow-hidden shadow-lg flex flex-col h-[280px]">
                  <div className="px-6 py-4 border-b border-quant-border bg-[#111827] flex justify-between items-center">
                    <h2 className="text-lg font-bold text-white tracking-wide">Main Allocation Equation</h2>
                  </div>
                  <div className="p-8 bg-black flex-1 flex flex-col justify-center">
                    <input 
                      type="text" 
                      value={equation}
                      onChange={(e) => setEquation(e.target.value)}
                      className="w-full bg-transparent text-blue-300 font-mono text-4xl focus:outline-none tracking-wider mb-6"
                    />
                    <button onClick={startSolving} disabled={isSolving} className="w-fit bg-quant-primary hover:bg-blue-600 text-white font-bold py-3 px-12 rounded-lg transition-all disabled:opacity-50 shadow-[0_0_15px_rgba(30,102,245,0.2)]">
                      {isSolving ? 'EXECUTING...' : 'Execute Solver'}
                    </button>
                  </div>
                </div>

                <div className="col-span-1 bg-[#0D1321] border border-quant-border rounded-xl p-6 shadow-lg flex flex-col h-[280px]">
                  <h3 className="text-xs font-bold text-white mb-4 uppercase tracking-widest text-quant-text">⚡ Execution Terminal</h3>
                  <div className="flex-1 bg-black rounded border border-quant-border p-4 font-mono text-[11px] overflow-y-auto space-y-2 custom-scrollbar">
                    {logs.length === 0 ? <span className="text-quant-text opacity-30">System idle. Awaiting command.</span> : logs.map((log, i) => (
                        <div key={i} className={log.includes('ERROR') ? 'text-rose-400' : 'text-blue-300'}>
                          <span className="text-gray-600 mr-2">{'>'}</span>{log}
                        </div>
                    ))}
                    {isSolving && <div className="text-quant-accent animate-pulse">_</div>}
                  </div>
                </div>
              </div>

              {/* RESULTS & RULE ENGINE */}
              <div className="grid grid-cols-3 gap-6 h-[400px]">
                <div className="col-span-2 bg-[#0D1321] border border-quant-border rounded-xl shadow-lg flex flex-col overflow-hidden">
                  <div className="px-6 py-4 border-b border-quant-border bg-[#111827] text-white font-bold text-sm tracking-wide">📊 Valid Solutions Found</div>
                  
                  <div className="flex gap-16 px-8 py-4 border-b border-quant-border bg-quant-bg/80 text-lg font-bold text-white">
                    {tableHeaders.length > 0 ? tableHeaders.map(header => (
                      <div key={header} className="w-32">{header}</div>
                    )) : <div className="text-quant-text text-base font-normal">No data streams active.</div>}
                  </div>

                  <div className="flex-1 overflow-y-auto bg-[#0D1321] custom-scrollbar">
                    {solutions.map((sol, index) => {
                      if (sol.startsWith("ERROR:")) return <div key={index} className="m-4 p-4 bg-rose-950/30 border border-rose-900 rounded text-rose-400 font-mono text-sm">⚠️ {sol.replace("ERROR: ", "")}</div>;
                      const valMap = {};
                      sol.split(", ").forEach(p => { const [k, v] = p.split("="); valMap[k] = v; });
                      return (
                        <div key={index} className="flex gap-16 px-8 py-4 border-b border-quant-border/40 hover:bg-[#111827] transition-colors font-mono text-lg text-gray-300">
                          {tableHeaders.map(header => <div key={header} className="w-32">{valMap[header] || '-'}</div>)}
                        </div>
                      );
                    })}
                    <div ref={solutionsEndRef} />
                  </div>
                </div>

                <div className="col-span-1 bg-[#0D1321] border border-quant-border rounded-xl shadow-lg flex flex-col overflow-hidden">
                   <div className="px-6 py-4 border-b border-quant-border bg-[#111827] flex justify-between items-center text-white font-bold text-sm tracking-wide">
                    ⚖ Rule Engine
                    <button onClick={() => setShowRuleForm(!showRuleForm)} className="w-6 h-6 rounded bg-quant-primary flex items-center justify-center text-white text-lg hover:bg-blue-600 transition-all shadow-sm">+</button>
                  </div>
                  <div className="p-4 space-y-3 flex-1 overflow-y-auto custom-scrollbar">
                     {showRuleForm && (
                       <div className="bg-quant-bg border border-quant-primary rounded p-4 space-y-3 animate-in slide-in-from-top-2 duration-300">
                         <div className="text-[10px] text-quant-text uppercase font-bold mb-1">Define Asset Rule</div>
                         <div className="flex gap-2">
                           <input type="text" placeholder="Var" value={newRuleVar} onChange={e => setNewRuleVar(e.target.value)} className="w-full bg-black border border-quant-border text-white text-xs p-2 rounded outline-none focus:border-quant-primary"/>
                           <select value={newRuleOp} onChange={e => setNewRuleOp(e.target.value)} className="bg-black border border-quant-border text-white text-xs p-2 rounded outline-none">
                             <option value=">">&gt;</option><option value="=">=</option><option value="<">&lt;</option>
                           </select>
                           <input type="number" placeholder="Val" value={newRuleVal} onChange={e => setNewRuleVal(e.target.value)} className="w-full bg-black border border-quant-border text-white text-xs p-2 rounded outline-none focus:border-quant-primary"/>
                         </div>
                         <button onClick={handleAddRule} className="w-full bg-quant-primary/20 hover:bg-quant-primary text-quant-primary hover:text-white text-xs font-bold py-2 rounded transition-all">Apply to Engine</button>
                       </div>
                     )}
                     {rules.map((rule) => (
                       <div key={rule.id} className="bg-black border border-quant-border rounded p-3 flex justify-between items-center group hover:border-quant-primary transition-all">
                          <span className="text-white font-mono text-sm tracking-wide">{rule.variable.toUpperCase()} {rule.operator} {rule.value}</span>
                          <button onClick={() => handleDeleteRule(rule.id)} className="text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity">🗑</button>
                       </div>
                     ))}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default App;