import { useState, useRef, useEffect } from 'react';

function App() {
  // --- STATE & LOGIC ---
  const [equation, setEquation] = useState("150a + 100b + 50c + 10d = 5000");
  const [solutions, setSolutions] = useState([]);
  const [isSolving, setIsSolving] = useState(false);
  const [activeTab, setActiveTab] = useState('solver');
  
  const [rules, setRules] = useState([]);
  const [showRuleForm, setShowRuleForm] = useState(false);
  const [newRuleVar, setNewRuleVar] = useState('');
  const [newRuleOp, setNewRuleOp] = useState('>');
  const [newRuleVal, setNewRuleVal] = useState('');

  const [logs, setLogs] = useState([]);
  const solutionsEndRef = useRef(null);
  const ws = useRef(null);

  useEffect(() => {
    solutionsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [solutions]);

  const handleAddRule = () => {
    if (!newRuleVar || !newRuleVal) return;
    const newRule = { id: Date.now(), variable: newRuleVar.toLowerCase().trim(), operator: newRuleOp, value: newRuleVal.trim() };
    setRules([...rules, newRule]);
    setNewRuleVar(''); setNewRuleVal(''); setShowRuleForm(false);
  };

  const handleDeleteRule = (id) => setRules(rules.filter(r => r.id !== id));
  const getConstraintsString = () => rules.map(r => `${r.variable}${r.operator}${r.value}`).join(', ');

  const startSolving = () => {
    if (!equation.includes("=")) {
      setSolutions(["ERROR: Oops! You forgot the '=' sign. Please specify a budget (e.g. '= 5000')."]);
      setLogs(["> ERROR: Syntax validation failed. Missing '=' parameter."]);
      return;
    }
    
    setSolutions([]);
    setLogs([]);
    setIsSolving(true);

    setTimeout(() => {
      const tokens = equation.split(/([+=\s])/).filter(t => t.trim() !== '');
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

    setTimeout(() => {
      ws.current = new WebSocket('ws://localhost:8080/ws');
      ws.current.onopen = () => { ws.current.send(JSON.stringify({ equation: equation, constraints: getConstraintsString() })); };
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

  // --- NEW: DYNAMIC TABLE PARSER ---
  // This automatically finds the variables (a, b, c, etc.) to create the table headers
  const getDynamicHeaders = () => {
    if (solutions.length === 0) return [];
    const validSol = solutions.find(s => !s.startsWith("ERROR"));
    if (!validSol) return [];
    return validSol.split(", ").map(pair => pair.split("=")[0]);
  };
  const tableHeaders = getDynamicHeaders();

  return (
    <div className="min-h-screen bg-quant-bg text-quant-white font-sans flex overflow-hidden">
      
      {/* SIDEBAR */}
      <aside className="w-64 bg-quant-bg border-r border-quant-border flex flex-col z-10">
        <div className="h-16 flex items-center px-6">
          <span className="text-xl font-bold tracking-tight text-quant-primary">Quant<span className="text-white">Solve</span></span>
        </div>
        <div className="px-6 py-4 flex items-center gap-3">
          <div className="w-10 h-10 rounded bg-quant-panel border border-quant-border flex items-center justify-center text-lg shadow-inner">👨‍💼</div>
          <div><div className="text-sm font-bold text-white">QuantSolve</div><div className="text-xs text-quant-text">High-Frequency Analyst</div></div>
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

      {/* MAIN CONTENT AREA */}
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
            
            {/* KPI CARDS */}
            <div className="grid grid-cols-4 gap-6">
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Search Space</div>
                <div className="text-3xl font-mono text-blue-300 font-medium">4.2M+</div>
              </div>
              <div className="bg-[#0D1321] border border-quant-border rounded-xl p-5 shadow-lg">
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-2">Solutions Found</div>
                <div className="text-3xl font-mono text-quant-accent font-medium">{solutions.length > 0 && !solutions[0].startsWith("ERROR") ? solutions.length : "0"}</div>
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

            {/* EQUATION & TIMELINE */}
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
                <h3 className="text-sm font-bold text-white mb-4 flex items-center gap-2"><span className="text-blue-400">⚡</span> Execution Terminal</h3>
                <div className="flex-1 bg-black rounded border border-quant-border p-4 font-mono text-[11px] overflow-y-auto space-y-2 relative">
                  {logs.length === 0 ? <span className="text-quant-text opacity-50 absolute top-4 left-4">System idle. Awaiting input.</span> : logs.map((log, i) => (
                      <div key={i} className={`${log.includes('ERROR') ? 'text-rose-400' : 'text-blue-300'}`}>
                        <span className="text-quant-text mr-2">{'>'}</span>{log}
                      </div>
                  ))}
                  {isSolving && <div className="text-quant-accent animate-pulse mt-2 inline-block">_</div>}
                </div>
              </div>
            </div>

            {/* ROW 3: SOLUTIONS & RULE ENGINE */}
            <div className="grid grid-cols-3 gap-6 h-[400px]">
              
              {/* THE NEW DATA GRID TABLE */}
              {/* THE NEW DATA GRID TABLE */}
              <div className="col-span-6 bg-[#0D1321] border border-quant-border rounded-xl shadow-lg flex flex-col overflow-hidden">
                <div className="px-6 py-4 border-b border-quant-border flex justify-between items-center bg-[#111827]">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <span className="text-quant-accent">📊</span> Valid Solutions Found
                  </h3>
                </div>
                
                {/* Dynamic Table Header (UPDATED: Larger text, fixed column widths) */}
                <div className="flex gap-16 px-8 py-4 border-b border-quant-border bg-quant-bg/200 text-lg font-bold text-white tracking-widest">
                  {tableHeaders.length > 0 ? tableHeaders.map(header => (
                    <div key={header} className="w-32">{header}</div>
                  )) : (
                    <div className="text-quant-text text-base">Awaiting Data...</div>
                  )}
                </div>

                {/* Dynamic Table Body (UPDATED: Larger text, matching gaps) */}
                <div className="flex-1 overflow-y-auto custom-scrollbar relative bg-[#0D1321]">
                  {solutions.length === 0 && !isSolving && logs.length === 0 && (
                    <div className="absolute inset-0 flex items-center justify-center text-quant-text font-mono text-base opacity-50">
                      Awaiting execution command...
                    </div>
                  )}
                  
                  <div>
                    {solutions.map((sol, index) => {
                      // Handle our polite errors safely
                      if (sol.startsWith("ERROR:")) {
                        return (
                          <div key={index} className="m-4 px-4 py-3 bg-rose-950/30 border border-rose-900/50 rounded-lg text-rose-400 font-mono text-base">
                            <span>⚠️</span> {sol.replace("ERROR: ", "")}
                          </div>
                        );
                      }

                      // Parse "a=40, b=0, c=10" into a dictionary like { a: "40", b: "0", c: "10" }
                      const pairs = sol.split(", ");
                      const valMap = {};
                      pairs.forEach(p => {
                        const [key, value] = p.split("=");
                        valMap[key] = value;
                      });

                      return (
                        <div key={index} className="flex gap-16 px-8 py-4 border-b border-quant-border/40 hover:bg-quant-bg/50 transition-colors font-mono text-lg items-center text-gray-300">
                          {tableHeaders.map(header => (
                            <div key={header} className="w-32">
                              {valMap[header] || '-'}
                            </div>
                          ))}
                        </div>
                      );
                    })}
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
                  <button onClick={() => setShowRuleForm(!showRuleForm)} className="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-bold hover:bg-blue-400 transition-colors">
                    {showRuleForm ? '×' : '+'}
                  </button>
                </div>
                
                <div className="p-4 space-y-3 flex-1 overflow-y-auto custom-scrollbar">
                   {showRuleForm && (
                     <div className="bg-quant-bg border border-quant-primary rounded-lg p-3">
                       <div className="text-[10px] text-quant-text mb-2 uppercase tracking-widest">Define Constraint</div>
                       <div className="flex gap-2 mb-3">
                         <input type="text" placeholder="Var" value={newRuleVar} onChange={e => setNewRuleVar(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono"/>
                         <select value={newRuleOp} onChange={e => setNewRuleOp(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono">
                           <option value=">">&gt;</option><option value=">=">&ge;</option><option value="<">&lt;</option><option value="<=">&le;</option><option value="=">=</option>
                         </select>
                         <input type="number" placeholder="Val" value={newRuleVal} onChange={e => setNewRuleVal(e.target.value)} className="w-1/3 bg-black border border-quant-border text-white text-xs px-2 py-1 rounded outline-none font-mono"/>
                       </div>
                       <button onClick={handleAddRule} className="w-full bg-quant-primary/20 hover:bg-quant-primary text-quant-primary hover:text-white text-xs font-bold py-1.5 rounded transition-colors">Add to Engine</button>
                     </div>
                   )}
                   {rules.map((rule) => (
                     <div key={rule.id} className="bg-black border border-quant-border rounded-lg p-3 group hover:border-quant-accent/50 transition-colors">
                        <div className="flex justify-between items-start mb-1">
                          <span className="font-mono text-sm text-white">Asset {rule.variable.toUpperCase()} {rule.operator} {rule.value}</span>
                          <button onClick={() => handleDeleteRule(rule.id)} className="text-quant-text hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity">🗑</button>
                        </div>
                     </div>
                   ))}
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