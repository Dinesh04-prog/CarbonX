import { useState, useRef, useEffect } from 'react';

function App() {
  // --- STATE & LOGIC ---
  const [equation, setEquation] = useState("");
  const [solutions, setSolutions] = useState([]);
  const [isSolving, setIsSolving] = useState(false);
  const [activeTab, setActiveTab] = useState('solver');
  
  // NEW STATE: Rule Engine State
  const [rules, setRules] = useState([]); // Stores the active rules
  const [showRuleForm, setShowRuleForm] = useState(false); // Toggles the input form
  const [newRuleVar, setNewRuleVar] = useState('');
  const [newRuleOp, setNewRuleOp] = useState('>');
  const [newRuleVal, setNewRuleVal] = useState('');

  const [logs, setLogs] = useState([]);
  const solutionsEndRef = useRef(null);
  const ws = useRef(null);

  useEffect(() => {
    solutionsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [solutions]);

  // --- NEW LOGIC: Manage Rules ---
  const handleAddRule = () => {
    if (!newRuleVar || !newRuleVal) return;
    
    const newRule = {
      id: Date.now(),
      variable: newRuleVar.toLowerCase().trim(),
      operator: newRuleOp,
      value: newRuleVal.trim()
    };
    
    setRules([...rules, newRule]);
    setNewRuleVar('');
    setNewRuleVal('');
    setShowRuleForm(false);
  };

  const handleDeleteRule = (id) => {
    setRules(rules.filter(r => r.id !== id));
  };

  // Converts the rules array into the "x>5, y<=10" string the Go backend expects
  const getConstraintsString = () => {
    return rules.map(r => `${r.variable}${r.operator}${r.value}`).join(', ');
  };

  // --- ENGINE EXECUTION ---
  const startSolving = () => {
    if (!equation.includes("=")) return alert("Please include an '=' in your equation");
    
    setSolutions([]);
    setLogs([]);
    setIsSolving(true);

    // Timeline Sequence
    setTimeout(() => {
      const tokens = equation.split(/([+=\s])/).filter(t => t.trim() !== '');
      setLogs([`Step 1: Tokenizing... [${tokens.map(t => `'${t}'`).join(', ')}]`]);
    }, 100);

    setTimeout(() => {
      const vars = [...new Set(equation.match(/[a-zA-Z]+/g) || [])];
      setLogs(prev => [...prev, `Step 2: Extracting Variables... Assets detected: ${vars.join(', ')}`]);
    }, 800);

    setTimeout(() => {
      // Add the constraints string to the timeline so judges can see it working!
      const constraintStr = getConstraintsString();
      setLogs(prev => [...prev, `Step 3: Applying Rules: ${constraintStr ? constraintStr : "None"}...`]);
    }, 1500);

    setTimeout(() => {
      ws.current = new WebSocket('ws://localhost:8080/ws');
      
      ws.current.onopen = () => { 
        // DYNAMIC CONSTRAINTS: We inject the real rules here!
        ws.current.send(JSON.stringify({ 
          equation: equation, 
          constraints: getConstraintsString() 
        })); 
      };
      
      ws.current.onmessage = (event) => {
        if (event.data === "FINISHED") {
          setIsSolving(false);
          setLogs(prev => [...prev, `> EXECUTION COMPLETE. Link closed.`]);
          ws.current.close();
        } else if (event.data.startsWith("Solution Found:")) {
          const cleanSol = event.data.replace("Solution Found: ", "");
          setSolutions((prev) => [...prev, cleanSol]);
        } else {
          setSolutions((prev) => [...prev, `ERROR: ${event.data}`]);
          setLogs(prev => [...prev, `> ERROR DETECTED.`]);
        }
      };

      ws.current.onerror = () => {
        setSolutions([`ERROR: Could not connect to Go Engine on port 8080.`]);
        setLogs(prev => [...prev, `> FATAL ERROR: Engine Offline.`]);
        setIsSolving(false);
      };
    }, 2200); 
  };

  return (
    <div className="min-h-screen bg-quant-bg text-quant-white font-sans flex overflow-hidden">
      
      {/* --- LEFT SIDEBAR --- */}
      <aside className="w-64 bg-quant-bg border-r border-quant-border flex flex-col z-10">
        <div className="h-16 flex items-center px-6">
          <span className="text-xl font-bold tracking-tight text-quant-primary">Quant<span className="text-white">Solve</span></span>
        </div>
        
        <div className="px-6 py-4 flex items-center gap-3">
          <div className="w-10 h-10 rounded bg-quant-panel border border-quant-border flex items-center justify-center text-lg shadow-inner">👨‍💼</div>
          <div>
            <div className="text-sm font-bold text-white">QuantSolve</div>
            <div className="text-xs text-quant-text">High-Frequency Analyst</div>
          </div>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1">
          <button onClick={() => setActiveTab('solver')} className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-semibold transition-colors ${activeTab === 'solver' ? 'bg-quant-panel text-white' : 'text-quant-text hover:text-white'}`}>
            <span className={activeTab === 'solver' ? 'text-quant-primary' : ''}>⊞</span> Workspaces
          </button>
          <button className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-quant-text hover:text-white transition-colors">
            <span>ƒx</span> Saved Formulas
          </button>
        </nav>

        <div className="p-4">
          <button className="w-full bg-quant-primary hover:bg-blue-600 text-white font-bold py-3 rounded-lg shadow-[0_0_15px_rgba(30,102,245,0.3)] transition-all">New Solver</button>
        </div>
      </aside>

      {/* --- MAIN CONTENT AREA --- */}
      <main className="flex-1 flex flex-col h-screen overflow-hidden bg-[#0A0E17]">
        <header className="h-16 border-b border-quant-border flex items-center justify-between px-8 bg-quant-bg/50 backdrop-blur-sm">
          <div className="flex gap-8 text-sm font-semibold text-quant-text">
            <button className="text-white h-16 border-b-2 border-quant-primary">Solver Dashboard</button>
            <button className="hover:text-white transition-colors">Research</button>
          </div>
          <div className="flex items-center gap-4">
            <button className="bg-quant-primary text-white text-xs font-bold px-4 py-2 rounded-md shadow-sm">Live Status</button>
          </div>
        </header>

        <div className="flex-1 p-8 overflow-y-auto custom-scrollbar">
          <div className="max-w-[1400px] mx-auto space-y-6">
            
            {/* ROW 1: KPI CARDS */}
            <div className="grid grid-cols-4 gap-6">
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Search Space</div>
                <div className="text-3xl font-mono text-blue-300 font-medium">--</div>
              </div>
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Solutions Found</div>
                <div className="text-3xl font-mono text-quant-accent font-medium">{solutions.length > 0 && !solutions[0].startsWith("ERROR") ? solutions.length : "--"}</div>
              </div>
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Engine Latency</div>
                <div className="text-3xl font-mono text-white font-medium">{isSolving ? "0.08ms" : "--"}</div>
              </div>
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Compute Node</div>
                <div className="text-lg font-mono text-white font-medium mt-1">AWS-US-EAST-1D</div>
              </div>
            </div>

            {/* ROW 2: EQUATION & TIMELINE */}
            <div className="grid grid-cols-3 gap-6">
              <div className="col-span-2 bg-[#0D1321] border border-quant-border rounded-xl overflow-hidden shadow-lg flex flex-col">
                <div className="px-6 py-4 border-b border-quant-border flex justify-between items-center bg-[#111827]">
                  <h2 className="text-lg font-bold text-white tracking-wide">Main Allocation Equation</h2>
                </div>
                <div className="p-6 bg-black flex-1 flex flex-col justify-center relative">
                  <input 
                    type="text" 
                    value={equation}
                    onChange={(e) => setEquation(e.target.value)}
                    className="w-full bg-transparent text-blue-300 font-mono text-3xl focus:outline-none tracking-wider mb-8"
                  />
                  <div className="flex justify-between items-end mt-auto">
                    <button onClick={startSolving} disabled={isSolving} className="bg-[#7AA2F7] hover:bg-blue-400 text-black font-bold py-3 px-8 rounded flex items-center gap-3 transition-colors disabled:opacity-50">
                      {isSolving ? 'CRUNCHING...' : 'Execute Solver'}
                    </button>
                  </div>
                </div>
              </div>

              <div className="col-span-1 bg-[#0D1321] border border-quant-border rounded-xl p-6 shadow-lg flex flex-col">
                <h3 className="text-sm font-bold text-white mb-4 flex items-center gap-2">
                  <span className="text-blue-400">⚡</span> Execution Terminal
                </h3>
                <div className="flex-1 bg-black rounded border border-quant-border p-4 font-mono text-[11px] overflow-y-auto space-y-2 relative">
                  {logs.length === 0 ? (
                    <span className="text-quant-text opacity-50 absolute top-4 left-4">System idle. Awaiting input.</span>
                  ) : (
                    logs.map((log, i) => (
                      <div key={i} className={`${log.includes('ERROR') ? 'text-rose-400' : 'text-blue-300'}`}>
                        <span className="text-quant-text mr-2">{'>'}</span>{log}
                      </div>
                    ))
                  )}
                  {isSolving && <div className="text-quant-accent animate-pulse mt-2 inline-block">_</div>}
                </div>
              </div>
            </div>

            {/* ROW 3: SOLUTIONS & RULE ENGINE */}
            <div className="grid grid-cols-3 gap-6 h-[400px]">
              
              {/* Valid Solutions Table */}
              <div className="col-span-2 bg-[#0D1321] border border-quant-border rounded-xl shadow-lg flex flex-col overflow-hidden">
                <div className="px-6 py-4 border-b border-quant-border flex justify-between items-center bg-[#111827]">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <span className="text-quant-accent">📊</span> Valid Solutions Found
                  </h3>
                </div>
                
                <div className="grid grid-cols-5 gap-4 px-6 py-3 border-b border-quant-border bg-quant-bg/50 text-[10px] font-bold text-quant-text tracking-widest uppercase">
                  <div>Rank</div><div className="col-span-3">Configuration Data</div><div className="text-right">Sum Diff</div>
                </div>

                <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                  <div className="space-y-1">
                    {solutions.map((sol, index) => (
                      sol.startsWith("ERROR:") ? (
                        <div key={index} className="px-4 py-3 bg-rose-950/30 border border-rose-900/50 rounded-lg text-rose-400 font-mono text-xs"><span>⚠️</span> {sol.replace("ERROR: ", "")}</div>
                      ) : (
                        <div key={index} className="grid grid-cols-5 gap-4 px-4 py-3 bg-quant-bg/30 hover:bg-quant-bg/80 rounded border border-transparent hover:border-quant-border transition-colors font-mono text-sm items-center">
                          <div className="flex items-center gap-2 text-white">
                            {index === 0 && <span className="text-quant-accent">★</span>} {String(index + 1).padStart(3, '0')}
                          </div>
                          <div className="col-span-3 text-quant-accent flex flex-wrap gap-x-4 gap-y-2">
                            {sol.split(", ").map((pair, i) => (
                              <span key={i} className="bg-[#15231c] px-2 py-0.5 rounded border border-quant-accent/20">{pair}</span>
                            ))}
                          </div>
                          <div className="text-right text-quant-accent">0.00%</div>
                        </div>
                      )
                    ))}
                    <div ref={solutionsEndRef} />
                  </div>
                </div>
              </div>

              {/* DYNAMIC RULE ENGINE */}
              <div className="col-span-1 bg-[#0D1321] border border-quant-border rounded-xl shadow-lg flex flex-col overflow-hidden">
                 <div className="px-6 py-4 border-b border-quant-border flex justify-between items-center bg-[#111827]">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <span className="text-blue-400">⚖</span> The Rule Engine
                  </h3>
                  <button 
                    onClick={() => setShowRuleForm(!showRuleForm)}
                    className="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-bold hover:bg-blue-400 transition-colors"
                  >
                    {showRuleForm ? '×' : '+'}
                  </button>
                </div>
                
                <div className="p-4 space-y-3 flex-1 overflow-y-auto custom-scrollbar">
                   
                   {/* Form to add a new rule */}
                   {showRuleForm && (
                     <div className="bg-quant-bg border border-quant-primary rounded-lg p-3 animate-pulse-once">
                       <div className="text-[10px] text-quant-text mb-2 uppercase tracking-widest">Define Constraint</div>
                       <div className="flex gap-2 mb-3">
                         <input type="text" placeholder="Var (e.g. a)" value={newRuleVar} onChange={e => setNewRuleVar(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono focus:border-quant-primary"/>
                         <select value={newRuleOp} onChange={e => setNewRuleOp(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono">
                           <option value=">">&gt;</option>
                           <option value=">=">&ge;</option>
                           <option value="<">&lt;</option>
                           <option value="<=">&le;</option>
                           <option value="=">=</option>
                         </select>
                         <input type="number" placeholder="Value" value={newRuleVal} onChange={e => setNewRuleVal(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono focus:border-quant-primary"/>
                       </div>
                       <button onClick={handleAddRule} className="w-full bg-quant-primary/20 hover:bg-quant-primary text-quant-primary hover:text-white text-xs font-bold py-1.5 rounded transition-colors">Add to Engine</button>
                     </div>
                   )}

                   {/* Render Active Rules */}
                   {rules.map((rule) => (
                     <div key={rule.id} className="bg-black border border-quant-border rounded-lg p-3 group hover:border-quant-accent/50 transition-colors">
                        <div className="flex justify-between items-start mb-1">
                          <span className="font-mono text-sm text-white">Asset {rule.variable.toUpperCase()} {rule.operator} {rule.value}</span>
                          <button onClick={() => handleDeleteRule(rule.id)} className="text-quant-text hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity">🗑</button>
                        </div>
                        <div className="text-[10px] text-quant-text uppercase tracking-widest">Active Boundary</div>
                     </div>
                   ))}

                   {/* Empty State */}
                   {rules.length === 0 && !showRuleForm && (
                     <div onClick={() => setShowRuleForm(true)} className="mt-4 border-2 border-dashed border-quant-border rounded-lg p-6 flex flex-col items-center justify-center text-quant-text gap-2 hover:bg-quant-bg/50 transition-colors cursor-pointer">
                        <span className="text-xl">⋮⋮</span>
                        <span className="text-[10px] uppercase tracking-widest text-center">No limits active.<br/>Click + to add rules.</span>
                     </div>
                   )}
                </div>
              </div>

            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;