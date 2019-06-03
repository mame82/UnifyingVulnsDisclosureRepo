using System;
using System.Text;
using System.IO;
using System.Threading;


namespace LogitackerClient
{

    class UnifyingUSB
    {
        public FileStream hidpp_short_file;
        public FileStream hidpp_long_file;
        public FileStream dj_long_file;

        public const UInt16 VID = 0x046d;
        public const UInt16 PID = 0xc52b;

        public const UInt16 HIDPP_SHORT_LENGTH = 7;
        public const UInt16 HIDPP_LONG_LENGTH = 20;
        public const UInt16 DJ_LONG_LENGTH = 32;

       
        public RProc rProc;

        public UnifyingUSB()
        {
            

            this.hidpp_short_file = Device.Open(VID, PID, HIDPP_SHORT_LENGTH);
            this.hidpp_long_file = Device.Open(VID, PID, HIDPP_LONG_LENGTH);
            this.dj_long_file = Device.Open(VID, PID, DJ_LONG_LENGTH);

        }

        public void BindProcess(bool withStdErr, string procName, string procArgs) {
            if (this.rProc != null) {
                //kill old rProc
            }
            this.rProc = new RProc(withStdErr,procName,procArgs);
        }

        public void RunShell(string procName, string procArgs) {
            this.BindProcess(true, procName, procArgs);

            byte inLastSeq = 3;
            byte outSeq = 0;

            byte[] outrep = new byte[HIDPP_LONG_LENGTH];;
            outrep[0] = 0x11;
            outrep[1] = 0x03;
            outrep[2] = 0xba;
            bool outIsControlFrame = false;
            byte outControlFrameType = 0;
            byte[] outPayload = new byte[0];
            byte outPayloadLength = 0;


            while (this.rProc.IsRunning()) {

                //byte[] inrep = uu.ReadHIDInReport(false);

                byte[] inrep = new byte[UnifyingUSB.HIDPP_LONG_LENGTH];
                int l = this.hidpp_long_file.Read(inrep, 0, inrep.Length);
  

                if (inrep.Length == 20 && (inrep[2] == 0xbb || inrep[2] == 0xba)) { //ToDo: replace with full frame validation
                    //Console.WriteLine(String.Format("In  {0}", Helper.ByteArrayToString(inrep)));

                    //copy over device IDX to outrep, to respond on proper RF address
                    outrep[1] = inrep[1];
                    
                    byte bitmaskIn = inrep[3];
                    byte inPaylen = (byte) ((bitmaskIn & 0xf0) >> 4);
                    byte inAck = (byte) ((bitmaskIn & 0x0c) >> 2);
                    byte inSeq = (byte) (bitmaskIn & 0x3) ;
                    byte inNextSeq = (byte) ((inLastSeq + 1) % 4);
                    bool inIsControlFrame = inrep[2] == 0xbb;
  
                    byte outAck = inLastSeq;

                    // is received report a new one ?
                    if (inSeq == inNextSeq) {
                        //New input frame (no re-transmit or invalid seq)
                        inLastSeq = inSeq;
                        outAck = inSeq;
                        //Console.WriteLine(String.Format("New input {0}", Helper.ByteArrayToString(inrep)));

                        if (inIsControlFrame && inPaylen == 0) { //paylen corresponds to control type, if control type bit is set; control type 0 is a frame with maximum payload length
                            inPaylen = 16;
                        }
                        
                        // we have to filter out packets with empty payload, which are sent in reply
                        // to update sequence numbers
                        if (inPaylen > 0) {
                            byte[] inPay = new byte[inPaylen];
                            Array.Copy(inrep, 4, inPay, 0, inPaylen);
                             Console.Write(String.Format("{0}", Encoding.UTF8.GetString(inPay)));
                             this.rProc.ToStdin(inPay);
                        }
                    }

                    //Last USB report received by device ??
                    if (inAck == outSeq) {
                        //Console.WriteLine("Last payload transmitted, ready for new one");
                        
                        outSeq = (byte) ((outSeq + 1) % 4);

                        //update payload, depending on pending data
                        if (this.rProc.HasOut()) {
                            outPayload = this.rProc.GetOut();
                            Console.Write(String.Format("{0}", Encoding.UTF8.GetString(outPayload)));
                            outPayloadLength = outPayload.Length > 16 ? (byte) 16 : (byte) outPayload.Length;

                            if (outPayloadLength == 16) {
                                outIsControlFrame = true;
                                outControlFrameType = 0x0;
                            } else {
                                outIsControlFrame = false;
                            }
                            
                        } else {
                            //Console.WriteLine("Proc has no data on stdout");
                            outIsControlFrame = false;
                            outPayloadLength = 0;
                        }
                    }


                    //update out report
                    if (outIsControlFrame) outrep[2] |= 0x01; //set control frame bit
                    else outrep[2] &= 0xfe; //unset control frame bit

                    byte bitmaskOut = outSeq;
                    bitmaskOut |= (byte) (outAck << 2);
                    bitmaskOut |= (byte) (outPayloadLength << 4);

                    if (outIsControlFrame) {
                        //replace length in bitmask field with control frame type
                        bitmaskOut &= 0x0f;
                        bitmaskOut |= (byte) (outControlFrameType << 4);


                        switch (outControlFrameType) {
                            case 0:
                                break;
                        }
                    } 
                    
                    outrep[3] = bitmaskOut;
                    Array.Copy(outPayload, 0, outrep, 4, outPayloadLength);
                    //this.WriteUSBOutputReport(outrep);
                    this.hidpp_long_file.Write(outrep, 0, outrep.Length);
                    this.hidpp_long_file.Flush();


                    //Console.WriteLine(String.Format("Out {0}", Helper.ByteArrayToString(outrep)));
                }

                // delay to avoid flooding USB report queues faster than RF is working and
                // keep room for real device communication
                Thread.Sleep(4);
            }

        }

    }
}
