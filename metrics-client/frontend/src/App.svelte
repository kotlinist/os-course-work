<script>
  import "./app.css";
  import logo from "./assets/images/logo-universal.png";
  // import { Greet } from "../wailsjs/go/main/App.js";
  import MousePanel from "../lib/MousePanel.svelte";
  import ProcessPanel from "../lib/ProcessPanel.svelte";

  let resultText = $state("Please enter your name below 👇");
  let name = $state();

  // function greet() {
  //   Greet(name).then((result) => (resultText = result));
  // }
</script>

<main>
  <!-- <button class="btn" onclick={greet}>Greet</button> -->
  <div class="min-h-screen flex flex-col flex-grow w-full h-full p-6 gap-4">
    <MousePanel />
    <ProcessPanel />
  </div>
</main>

<style>
  #logo {
    display: block;
    width: 50%;
    height: 50%;
    margin: auto;
    padding: 10% 0 0;
    background-position: center;
    background-repeat: no-repeat;
    background-size: 100% 100%;
    background-origin: content-box;
  }

  .result {
    height: 20px;
    line-height: 20px;
    margin: 1.5rem auto;
  }

  .input-box .btn {
    width: 60px;
    height: 30px;
    line-height: 30px;
    border-radius: 3px;
    border: none;
    margin: 0 0 0 20px;
    padding: 0 8px;
    cursor: pointer;
  }

  .input-box .btn:hover {
    background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
    color: #333333;
  }

  .input-box .input {
    border: none;
    border-radius: 3px;
    outline: none;
    height: 30px;
    line-height: 30px;
    padding: 0 10px;
    background-color: rgba(240, 240, 240, 1);
    -webkit-font-smoothing: antialiased;
  }

  .input-box .input:hover {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .input-box .input:focus {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }
</style>
