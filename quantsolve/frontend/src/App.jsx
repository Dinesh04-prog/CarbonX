import { useState, useRef } from 'react';

function App() {
  const [equation, setEquation] = useState("10x + 20y = 100");
  const [solutions, setSolutions] = useState([]);
  const [isSolving, setIsSolving] = useState(false);
  const ws = useRef(null);

  const startSolving = () => {
    if (!equation.includes("=")) return alert("Please include an '=' in your equation");
    
    setSolutions([]);
    setIsSolving(true);

    ws.current = new WebSocket('ws://localhost:8080/ws');

    ws.current.onopen = () => {
      ws.current.send(equation);
    };

    ws.current.onmessage = (event) => {
      if (event.data === "FINISHED") {
        setIsSolving(false);
        ws.current.close();
      } else {
        setSolutions((prev) => [...prev, event.data]);
      }
    };

    ws.current.onerror = () => {
      setSolutions(["Error: Could not connect to backend engine."]);
      setIsSolving(false);
    };
  };

  return (
    <div className="min-h-screen bg-slate-900 text-white p-8 font-mono">
      <div className="max-w-3xl mx-auto">
        <header className="mb-12 border-b border-slate-800 pb-6">
          <h1 className="text-4xl font-black text-emerald-400 tracking-tighter">QUANTSOLVE_</h1>
          <p className="text-slate-400 mt-2">Proprietary Algebraic Allocation Engine v1.0</p>
        </header>

        <div className="bg-slate-800 p-8 rounded-xl shadow-2xl border border-slate-700 mb-10">
          <label className="block text-xs font-bold uppercase text-slate-500 mb-3 tracking-widest">
            Enter Market Equation
          </label>
          <div className="flex gap-4">
            <input 
              type="text" 
              value={equation}
              onChange={(e) => setEquation(e.target.value)}
              placeholder="e.g. 50x + 10y = 500"
              className="flex-1 bg-slate-950 border border-slate-600 rounded-lg px-5 py-3 text-emerald-400 text-lg focus:ring-2 focus:ring-emerald-500 focus:outline-none transition-all"
            />
            <button 
              onClick={startSolving}
              disabled={isSolving}
              className="bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-black py-3 px-10 rounded-lg transition-transform active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSolving ? 'CRUNCHING...' : 'SOLVE'}
            </button>
          </div>
        </div>

        <div className="bg-black rounded-xl border border-slate-800 overflow-hidden shadow-inner">
          <div className="bg-slate-800/50 px-4 py-2 border-b border-slate-800 flex justify-between items-center">
            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-tight">Real-Time Solution Stream</span>
            <div className={`h-2 w-2 rounded-full ${isSolving ? 'bg-emerald-500 animate-pulse' : 'bg-slate-700'}`}></div>
          </div>
          <div className="p-6 min-h-[400px] max-h-[500px] overflow-y-auto custom-scrollbar">
            {solutions.length === 0 && !isSolving && (
              <div className="text-slate-700 text-center mt-20 font-light italic">Waiting for market data...</div>
            )}
            <div className="space-y-2">
              {solutions.map((sol, index) => (
                <div key={index} className={`flex items-center gap-4 ${sol.includes("Error") ? 'text-rose-400' : 'text-emerald-300'}`}>
                  <span className="text-slate-600 text-xs w-8 tabular-nums">{String(index + 1).padStart(2, '0')}</span>
                  <span className="font-medium tracking-tight underline decoration-slate-800 underline-offset-4">{sol}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;