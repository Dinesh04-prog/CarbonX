import { useState, useRef, useEffect } from 'react';

function App() {
  // --- STATE & LOGIC (Unchanged) ---
  const [equation, setEquation] = useState("");
  const [solutions, setSolutions] = useState([]);
  const [isSolving, setIsSolving] = useState(false);
  const [activeTab, setActiveTab] = useState('solver');
  
  const solutionsEndRef = useRef(null);
  const ws = useRef(null);

  useEffect(() => {
    solutionsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [solutions]);

  const startSolving = () => {
    if (!equation.includes("=")) return alert("Please include an '=' in your equation");
    
    setSolutions([]);
    setIsSolving(true);
    ws.current = new WebSocket('ws://localhost:8080/ws');
    
    ws.current.onopen = () => { 
      ws.current.send(JSON.stringify({ equation: equation, constraints: "" })); 
    };
    
    ws.current.onmessage = (event) => {
      if (event.data === "FINISHED") {
        setIsSolving(false);
        ws.current.close();
      } else if (event.data.startsWith("Solution Found:")) {
        const cleanSol = event.data.replace("Solution Found: ", "");
        setSolutions((prev) => [...prev, cleanSol]);
      } else {
        setSolutions((prev) => [...prev, `ERROR: ${event.data}`]);
      }
    };

    ws.current.onerror = () => {
      setSolutions([`ERROR: Could not connect to Go Engine on port 8080. Is it running?`]);
      setIsSolving(false);
    };
  };

  return (
    // 1. ADDED A SMOOTH BACKGROUND GRADIENT
    <div className="min-h-screen bg-gradient-to-br from-[#0B0F19] via-[#131B2C] to-[#090D16] text-quant-white font-sans flex overflow-hidden">
      
      {/* --- LEFT SIDEBAR (Glassy) --- */}
      <aside className="w-64 bg-white/5 backdrop-blur-xl border-r border-white/10 flex flex-col z-10 shadow-2xl relative">
        <div className="h-16 flex items-center px-6">
          <span className="text-xl font-bold tracking-tight text-blue-400 drop-shadow-[0_0_10px_rgba(59,130,246,0.5)]">
            Quant<span className="text-white">Solve</span>
          </span>
        </div>
        
        <div className="px-6 py-6 flex items-center gap-4">
          <div className="w-12 h-12 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500 p-0.5 shadow-lg">
             <div className="w-full h-full rounded-full bg-[#131B2C] flex items-center justify-center text-xl">👨‍💼</div>
          </div>
          <div>
            <div className="text-sm font-bold text-white">QuantSolve</div>
            <div className="text-xs text-blue-300/80">High-Frequency</div>
          </div>
        </div>

        <nav className="flex-1 px-4 py-4 space-y-2">
          {/* Rounded-xl for smoother tabs */}
          <button 
            onClick={() => setActiveTab('solver')}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all ${activeTab === 'solver' ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 text-white border border-white/10 shadow-[inset_0_1px_1px_rgba(255,255,255,0.1)]' : 'text-quant-text hover:text-white hover:bg-white/5'}`}
          >
            <span className={activeTab === 'solver' ? 'text-blue-400' : ''}>⊞</span> Workspaces
          </button>
          <button className="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-medium text-quant-text hover:text-white hover:bg-white/5 transition-all">
            <span>ƒx</span> Saved Formulas
          </button>
          <button className="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-medium text-quant-text hover:text-white hover:bg-white/5 transition-all">
            <span>⏱</span> Rules History
          </button>
        </nav>

        <div className="p-6">
          {/* Glossy Pill Button */}
          <button className="w-full bg-gradient-to-r from-blue-500 to-indigo-500 hover:from-blue-400 hover:to-indigo-400 text-white font-bold py-3.5 rounded-full shadow-[0_8px_20px_rgba(59,130,246,0.3)] hover:shadow-[0_8px_25px_rgba(59,130,246,0.5)] transition-all transform hover:-translate-y-0.5 border border-white/20">
            + New Solver
          </button>
        </div>
      </aside>

      {/* --- MAIN CONTENT AREA --- */}
      <main className="flex-1 flex flex-col h-screen overflow-hidden relative">
        
        {/* Glow Effects in the background */}
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/10 rounded-full blur-[120px] pointer-events-none"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-500/10 rounded-full blur-[120px] pointer-events-none"></div>

        {/* Top Navigation (Frosted Glass) */}
        <header className="h-16 border-b border-white/10 flex items-center justify-between px-8 bg-black/20 backdrop-blur-md z-10">
          <div className="flex gap-8 text-sm font-semibold text-quant-text">
            <button className="text-white h-16 border-b-2 border-blue-400 shadow-[0_2px_10px_rgba(96,165,250,0.5)]">Solver Dashboard</button>
            <button className="hover:text-white transition-colors">Research</button>
            <button className="hover:text-white transition-colors">Backtest</button>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-quant-text text-lg flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-quant-text shadow-[0_0_5px_rgba(255,255,255,0.3)]"></span>
              <span className="w-4 h-4 rounded-full bg-quant-text/50"></span>
            </span>
            {/* Pill shaped status indicator */}
            <button className="bg-white/10 border border-white/20 backdrop-blur text-white text-xs font-bold px-5 py-2 rounded-full shadow-lg">Live Status</button>
          </div>
        </header>

        {/* --- SCROLLABLE WORKSPACE --- */}
        <div className="flex-1 p-8 overflow-y-auto custom-scrollbar z-10 relative">
          <div className="max-w-[1400px] mx-auto space-y-8">
            
            {/* ROW 1: KPI CARDS (Glassmorphism) */}
            <div className="grid grid-cols-4 gap-6">
              <div className="bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl p-6 shadow-2xl relative overflow-hidden group hover:bg-white/[0.05] transition-all">
                <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-500 to-transparent opacity-50"></div>
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-3">Search Space</div>
                <div className="text-4xl font-mono text-white font-medium drop-shadow-md">4.2M+</div>
                <div className="w-full bg-black/40 h-1.5 mt-5 rounded-full overflow-hidden shadow-inner">
                  <div className="bg-gradient-to-r from-blue-400 to-indigo-500 w-3/4 h-full rounded-full shadow-[0_0_10px_rgba(96,165,250,0.8)]"></div>
                </div>
              </div>

              <div className="bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl p-6 shadow-2xl relative overflow-hidden group hover:bg-white/[0.05] transition-all">
                <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-emerald-500 to-transparent opacity-50"></div>
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-3">Solutions Found</div>
                <div className="text-4xl font-mono text-emerald-400 font-medium drop-shadow-[0_0_12px_rgba(52,211,153,0.3)]">
                  {solutions.length > 0 && !solutions[0].startsWith("ERROR") ? solutions.length : "1,024"}
                </div>
                <div className="text-xs text-emerald-400/80 mt-4 tracking-wide font-medium bg-emerald-500/10 inline-block px-3 py-1 rounded-full border border-emerald-500/20">+12.4% vs last</div>
              </div>

              <div className="bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl p-6 shadow-2xl relative overflow-hidden group hover:bg-white/[0.05] transition-all">
                 <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 to-transparent opacity-50"></div>
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-3">Engine Latency</div>
                <div className="text-4xl font-mono text-white font-medium drop-shadow-md">{isSolving ? "0.08ms" : "--"}</div>
                <div className="text-xs text-quant-text mt-4 tracking-wide bg-white/5 inline-block px-3 py-1 rounded-full border border-white/5">Real-time Stream</div>
              </div>

              <div className="bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl p-6 shadow-2xl relative overflow-hidden group hover:bg-white/[0.05] transition-all">
                 <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-amber-500 to-transparent opacity-50"></div>
                <div className="text-[10px] text-quant-text font-bold tracking-widest uppercase mb-3">Compute Node</div>
                <div className="text-xl font-mono text-white font-medium mt-1">AWS-US-EAST</div>
                <div className="flex items-center gap-2 text-xs font-bold text-amber-400 mt-4 tracking-wide bg-amber-500/10 inline-flex px-3 py-1 rounded-full border border-amber-500/20">
                  <span className={`w-2 h-2 rounded-full ${isSolving ? 'bg-amber-400 animate-pulse shadow-[0_0_8px_rgba(251,191,36,0.8)]' : 'bg-quant-text'}`}></span> {isSolving ? 'Active' : 'Idle'}
                </div>
              </div>
            </div>

            {/* ROW 2: MAIN EQUATION INPUT */}
<div className="col-span-2 bg-white/[0.03] backdrop-blur-2xl border border-white/10 rounded-3xl overflow-hidden shadow-2xl flex flex-col">
                <div className="px-8 py-5 border-b border-white/5 flex justify-between items-center bg-white/[0.02]">
                  <h2 className="text-lg font-bold text-white tracking-wide flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-400 border border-blue-500/30">ƒx</span> 
                    Main Allocation Equation
                  </h2>
                  <div className="flex gap-3">
                    <span className="text-[10px] px-3 py-1.5 rounded-full bg-black/40 text-quant-text border border-white/10 tracking-widest shadow-inner">ALGEBRAIC_V1</span>
                  </div>
                </div>
                
                {/* CHANGED: Removed justify-center and used gap-6 to keep elements tight */}
                <div className="p-8 bg-gradient-to-b from-transparent to-black/20 flex-1 flex flex-col gap-6 relative">
                  
                  <input 
                    type="text" 
                    value={equation}
                    onChange={(e) => setEquation(e.target.value)}
                    // CHANGED: Removed mb-8 so it doesn't push the button away
                    className="w-full bg-black/40 border border-white/10 rounded-2xl p-6 text-blue-300 font-mono text-4xl focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 focus:outline-none tracking-wider shadow-inner transition-all"
                  />
                  
                  {/* CHANGED: Removed mt-auto and changed to items-center */}
                  <div className="flex justify-between items-center">
                    <div className="flex gap-5 text-xs font-bold text-quant-text bg-black/30 px-5 py-2.5 rounded-full border border-white/5">
                      <span className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"></span> Constants</span>
                      <span className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-slate-400"></span> Variables</span>
                      <span className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.8)]"></span> Equality</span>
                    </div>
                    
                    <button 
                      onClick={startSolving}
                      disabled={isSolving}
                      className="bg-gradient-to-r from-emerald-400 to-teal-500 hover:from-emerald-300 hover:to-teal-400 text-slate-900 font-black py-4 px-10 rounded-full flex items-center gap-3 transition-all transform hover:scale-105 disabled:opacity-50 shadow-[0_10px_25px_rgba(52,211,153,0.4)] border border-white/40"
                    >
                      {isSolving ? (
                        <><span className="w-5 h-5 border-4 border-slate-900/30 border-t-slate-900 rounded-full animate-spin"></span> CRUNCHING...</>
                      ) : 'Execute Solver'}
                    </button>
                  </div>
                </div>
              </div>

            {/* --- NEW SECTION: ROW 3 --- */}
            <div className="grid grid-cols-3 gap-6 h-[450px]">
              
              {/* Left Side: Valid Solutions Table */}
              <div className="col-span-2 bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl shadow-2xl flex flex-col overflow-hidden">
                <div className="px-8 py-5 border-b border-white/5 flex justify-between items-center bg-white/[0.02]">
                  <h3 className="text-base font-bold text-white flex items-center gap-3">
                    <span className="p-2 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 shadow-[0_0_15px_rgba(52,211,153,0.2)]">📊</span> 
                    Valid Solutions Stream
                  </h3>
                  <div className="flex gap-2 bg-black/40 p-1 rounded-full border border-white/5 shadow-inner">
                    <button className="text-xs px-4 py-1.5 bg-white/10 text-white rounded-full shadow-sm">CSV</button>
                    <button className="text-xs px-4 py-1.5 text-quant-text rounded-full hover:text-white transition">JSON</button>
                  </div>
                </div>
                
                <div className="grid grid-cols-5 gap-4 px-8 py-4 border-b border-white/5 bg-black/20 text-xs font-bold text-quant-text tracking-widest uppercase shadow-inner">
                  <div>Rank</div>
                  <div className="col-span-3">Configuration Data</div>
                  <div className="text-right">Sum Diff</div>
                </div>

                <div className="flex-1 overflow-y-auto custom-scrollbar p-4">
                  {solutions.length === 0 && !isSolving && (
                    <div className="flex flex-col items-center justify-center h-full text-quant-text font-mono text-sm opacity-50 gap-4">
                      <span className="text-4xl text-white/20">⌘</span>
                      Awaiting execution command...
                    </div>
                  )}
                  
                  <div className="space-y-2">
                    {solutions.map((sol, index) => (
                      sol.startsWith("ERROR:") ? (
                        <div key={index} className="px-6 py-4 bg-gradient-to-r from-rose-950/80 to-rose-900/40 border border-rose-500/30 rounded-2xl text-rose-300 font-mono text-sm flex items-center gap-4 shadow-[0_5px_15px_rgba(225,29,72,0.2)] backdrop-blur-md">
                          <span className="text-xl p-1 bg-rose-500/20 rounded-full border border-rose-500/50">⚠️</span> {sol.replace("ERROR: ", "")}
                        </div>
                      ) : (
                        <div key={index} className="grid grid-cols-5 gap-4 px-6 py-4 bg-white/[0.02] hover:bg-white/[0.06] rounded-2xl border border-white/5 hover:border-white/20 transition-all font-mono text-sm items-center cursor-default transform hover:scale-[1.01] shadow-sm">
                          <div className="flex items-center gap-3 text-white">
                            {index === 0 ? <span className="text-amber-400 drop-shadow-[0_0_8px_rgba(251,191,36,0.8)]">★</span> : <span className="w-3.5 h-3.5 rounded-full border border-white/20"></span>}
                            {String(index + 1).padStart(3, '0')}
                          </div>
                          <div className="col-span-3 flex flex-wrap gap-x-3 gap-y-2">
                            {sol.split(", ").map((pair, i) => (
                              <span key={i} className="bg-blue-500/10 text-blue-300 px-3 py-1 rounded-full border border-blue-500/20 shadow-[inset_0_1px_2px_rgba(255,255,255,0.1)]">
                                {pair}
                              </span>
                            ))}
                          </div>
                          <div className="text-right text-emerald-400 font-bold bg-emerald-500/10 px-3 py-1 rounded-full justify-self-end border border-emerald-500/20">0.00%</div>
                        </div>
                      )
                    ))}
                    <div ref={solutionsEndRef} />
                  </div>
                </div>
              </div>

              {/* Right Side: The Rule Engine */}
              <div className="col-span-1 bg-white/[0.03] backdrop-blur-xl border border-white/10 rounded-3xl shadow-2xl flex flex-col overflow-hidden relative">
                 <div className="px-8 py-5 border-b border-white/5 flex justify-between items-center bg-white/[0.02]">
                  <h3 className="text-base font-bold text-white flex items-center gap-3">
                    <span className="p-2 rounded-full bg-purple-500/20 text-purple-400 border border-purple-500/30 shadow-[0_0_15px_rgba(168,85,247,0.2)]">⚖</span> 
                    Rule Engine
                  </h3>
                  <button className="w-8 h-8 rounded-full bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center text-white text-lg font-bold hover:shadow-[0_0_15px_rgba(168,85,247,0.5)] transition-all transform hover:scale-110 border border-white/20">
                    +
                  </button>
                </div>
                
                <div className="p-6 space-y-4 flex-1 overflow-y-auto custom-scrollbar">
                   {/* Dummy Rule 1 */}
                   <div className="bg-black/30 border border-white/10 rounded-2xl p-4 group hover:border-white/30 hover:bg-black/50 transition-all shadow-inner">
                      <div className="flex justify-between items-start mb-2">
                        <span className="font-mono text-base text-white bg-white/5 px-2 py-0.5 rounded border border-white/5">Asset A &gt; 0</span>
                        <button className="text-quant-text hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity bg-white/5 p-1.5 rounded-full">🗑</button>
                      </div>
                      <div className="text-[11px] text-quant-text uppercase tracking-widest font-semibold flex items-center gap-2">
                         <span className="w-1.5 h-1.5 rounded-full bg-purple-400"></span> Non-Negative Floor
                      </div>
                   </div>
                   
                   {/* Dummy Rule 2 */}
                   <div className="bg-black/30 border border-white/10 rounded-2xl p-4 group hover:border-white/30 hover:bg-black/50 transition-all shadow-inner">
                      <div className="flex justify-between items-start mb-2">
                        <span className="font-mono text-base text-white bg-white/5 px-2 py-0.5 rounded border border-white/5">Asset C % 2 = 0</span>
                        <button className="text-quant-text hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity bg-white/5 p-1.5 rounded-full">🗑</button>
                      </div>
                      <div className="text-[11px] text-quant-text uppercase tracking-widest font-semibold flex items-center gap-2">
                         <span className="w-1.5 h-1.5 rounded-full bg-purple-400"></span> Even-Weighted Check
                      </div>
                   </div>

                   {/* Drop Zone */}
                   <div className="mt-6 border-2 border-dashed border-white/10 rounded-3xl p-8 flex flex-col items-center justify-center text-quant-text gap-3 hover:bg-white/[0.02] hover:border-white/30 transition-all cursor-pointer bg-black/20">
                      <span className="text-3xl text-white/20">📥</span>
                      <span className="text-xs uppercase tracking-widest font-bold">Drop Rules Here</span>
                   </div>
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